package sunrisesunset

import (
	"context"
	"fmt"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes sunrise-sunset as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/sunrise-sunset-cli/sunrise-sunset"
//
// The init below registers it; the host then dereferences sunrise-sunset://
// URIs by routing to the operations Register installs. The same Domain also
// builds the standalone sunrise-sunset binary.
func init() { kit.Register(Domain{}) }

// Domain is the sunrise-sunset driver. It carries no state; the per-run client
// is built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "sunrisesunset",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "sunrise-sunset",
			Short:  "A command line for sunrise-sunset.org.",
			Long: `A command line for sunrise-sunset.org.

sunrise-sunset reads public solar data from api.sunrise-sunset.org over plain
HTTPS, shapes it into clean records, and prints output that pipes into the rest
of your tools. No API key, nothing to run alongside it.`,
			Site: "sunrise-sunset.org",
			Repo: "https://github.com/tamnd/sunrise-sunset-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// solar: fetch solar times for a lat/lng + optional date.
	kit.Handle(app, kit.OpMeta{
		Name:     "solar",
		Group:    "read",
		Single:   true,
		Summary:  "Get sunrise/sunset times for a location",
		URIType:  "point",
		Resolver: true,
	}, getSolar)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := NewClient()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.HTTP.Timeout = cfg.Timeout
	}
	return c, nil
}

// solarInput holds the flags for the solar command.
type solarInput struct {
	Lat    float64 `kit:"flag" help:"latitude"  default:"40.7128"`
	Lng    float64 `kit:"flag" help:"longitude" default:"-74.0060"`
	Date   string  `kit:"flag" help:"date YYYY-MM-DD (default: today)"`
	Client *Client `kit:"inject"`
}

func getSolar(ctx context.Context, in solarInput, emit func(*SolarInfo) error) error {
	info, err := in.Client.GetSolar(ctx, in.Lat, in.Lng, in.Date)
	if err != nil {
		return err
	}
	return emit(info)
}

// Classify turns an input string into a (type, id) pair.
// "lat,lon" or a bare coordinate pair maps to ("point", input).
// "YYYY-MM-DD" maps to ("date", input).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if len(input) == 10 && input[4] == '-' && input[7] == '-' {
		return "date", input, nil
	}
	if input == "" {
		return "", "", errs.Usage("provide a lat,lng pair or a YYYY-MM-DD date")
	}
	return "point", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "point":
		return fmt.Sprintf("https://sunrise-sunset.org/search?lat=%s&lng=%s", id, id), nil
	case "date":
		return "https://sunrise-sunset.org/search", nil
	default:
		return "", errs.Usage("sunrise-sunset has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind with the right exit
// code.
func mapErr(err error) error {
	return err
}
