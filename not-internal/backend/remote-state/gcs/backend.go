// Package gcs implements remote storage of state on Google Cloud Storage (GCS).
package gcs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/muratcelep/terraform/not-internal/backend"
	"github.com/muratcelep/terraform/not-internal/httpclient"
	"github.com/muratcelep/terraform/not-internal/legacy/helper/schema"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

// Backend implements "backend".Backend for GCS.
// Input(), Validate() and Configure() are implemented by embedding *schema.Backend.
// State(), DeleteState() and States() are implemented explicitly.
type Backend struct {
	*schema.Backend

	storageClient  *storage.Client
	storageContext context.Context

	bucketName string
	prefix     string

	encryptionKey []byte
}

func New() backend.Backend {
	b := &Backend{}
	b.Backend = &schema.Backend{
		ConfigureFunc: b.configure,
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Google Cloud Storage bucket",
			},

			"prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The directory where state files will be saved inside the bucket",
			},

			"credentials": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Google Cloud JSON Account Key",
				Default:     "",
			},

			"access_token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_OAUTH_ACCESS_TOKEN",
				}, nil),
				Description: "An OAuth2 token used for GCP authentication",
			},

			"impersonate_service_account": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_IMPERSONATE_SERVICE_ACCOUNT",
				}, nil),
				Description: "The service account to impersonate for all Google API Calls",
			},

			"impersonate_service_account_delegates": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The delegation chain for the impersonated service account",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"encryption_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A 32 byte base64 encoded 'customer supplied encryption key' used to encrypt all state.",
				Default:     "",
			},
		},
	}

	return b
}

func (b *Backend) configure(ctx context.Context) error {
	if b.storageClient != nil {
		return nil
	}

	// ctx is a background context with the backend config added.
	// Since no context is passed to remoteClient.Get(), .Lock(), etc. but
	// one is required for calling the GCP API, we're holding on to this
	// context here and re-use it later.
	b.storageContext = ctx

	data := schema.FromContextBackendConfig(b.storageContext)

	b.bucketName = data.Get("bucket").(string)
	b.prefix = strings.TrimLeft(data.Get("prefix").(string), "/")
	if b.prefix != "" && !strings.HasSuffix(b.prefix, "/") {
		b.prefix = b.prefix + "/"
	}

	var opts []option.ClientOption
	var credOptions []option.ClientOption

	// Add credential source
	var creds string
	var tokenSource oauth2.TokenSource

	if v, ok := data.GetOk("access_token"); ok {
		tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: v.(string),
		})
	} else if v, ok := data.GetOk("credentials"); ok {
		creds = v.(string)
	} else if v := os.Getenv("GOOGLE_BACKEND_CREDENTIALS"); v != "" {
		creds = v
	} else {
		creds = os.Getenv("GOOGLE_CREDENTIALS")
	}

	if tokenSource != nil {
		credOptions = append(credOptions, option.WithTokenSource(tokenSource))
	} else if creds != "" {

		// to mirror how the provider works, we accept the file path or the contents
		contents, err := backend.ReadPathOrContents(creds)
		if err != nil {
			return fmt.Errorf("Error loading credentials: %s", err)
		}

		if !json.Valid([]byte(contents)) {
			return fmt.Errorf("the string provided in credentials is neither valid json nor a valid file path")
		}

		credOptions = append(credOptions, option.WithCredentialsJSON([]byte(contents)))
	}

	// Service Account Impersonation
	if v, ok := data.GetOk("impersonate_service_account"); ok {
		ServiceAccount := v.(string)
		var delegates []string

		if v, ok := data.GetOk("impersonate_service_account_delegates"); ok {
			d := v.([]interface{})
			if len(delegates) > 0 {
				delegates = make([]string, len(d))
			}
			for _, delegate := range d {
				delegates = append(delegates, delegate.(string))
			}
		}

		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: ServiceAccount,
			Scopes:          []string{storage.ScopeReadWrite},
			Delegates:       delegates,
		}, credOptions...)

		if err != nil {
			return err
		}

		opts = append(opts, option.WithTokenSource(ts))

	} else {
		opts = append(opts, credOptions...)
	}

	opts = append(opts, option.WithUserAgent(httpclient.UserAgentString()))
	client, err := storage.NewClient(b.storageContext, opts...)
	if err != nil {
		return fmt.Errorf("storage.NewClient() failed: %v", err)
	}

	b.storageClient = client

	key := data.Get("encryption_key").(string)
	if key == "" {
		key = os.Getenv("GOOGLE_ENCRYPTION_KEY")
	}

	if key != "" {
		kc, err := backend.ReadPathOrContents(key)
		if err != nil {
			return fmt.Errorf("Error loading encryption key: %s", err)
		}

		// The GCS client expects a customer supplied encryption key to be
		// passed in as a 32 byte long byte slice. The byte slice is base64
		// encoded before being passed to the API. We take a base64 encoded key
		// to remain consistent with the GCS docs.
		// https://cloud.google.com/storage/docs/encryption#customer-supplied
		// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/def681/storage/storage.go#L1181
		k, err := base64.StdEncoding.DecodeString(kc)
		if err != nil {
			return fmt.Errorf("Error decoding encryption key: %s", err)
		}
		b.encryptionKey = k
	}

	return nil
}
