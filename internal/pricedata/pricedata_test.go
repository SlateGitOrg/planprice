package pricedata

import (
	"testing"

	"github.com/archit2501/planprice/internal/pricing"
)

func TestEmbeddedSnapshotMatchesChecksum(t *testing.T) {
	sum, err := VerifyEmbedded()
	if err != nil {
		t.Fatalf("embedded snapshot fails verification (rebuild with `make snapshot`, never edit by hand): %v", err)
	}
	if len(sum) != 64 {
		t.Fatalf("digest %q", sum)
	}
}

func TestEmbeddedSnapshotCoverage(t *testing.T) {
	snap, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	again, err := Default()
	if err != nil || again != snap {
		t.Fatal("Default should be decoded once and shared")
	}
	services := map[string]bool{}
	for _, o := range snap.Offers {
		services[o.Service+"/"+o.Region] = true
		if o.Products == 0 {
			t.Errorf("offer %s/%s contributed no prices", o.Service, o.Region)
		}
	}
	for _, region := range []string{"us-east-1", "eu-west-1"} {
		for _, svc := range []string{"AmazonEC2", "AmazonRDS", "AmazonS3", "AWSLambda", "AWSELB"} {
			if !services[svc+"/"+region] {
				t.Errorf("snapshot lacks %s in %s", svc, region)
			}
		}
	}
	// Spot-check one well-known price per service so a bad rebuild is caught.
	checks := []struct {
		q    pricing.Query
		want string
	}{
		{pricing.Query{Service: "AmazonEC2", Region: "us-east-1", Family: "Compute Instance", Attrs: map[string]string{"instanceType": "t3.micro", "operatingSystem": "Linux"}}, "0.0104"},
		{pricing.Query{Service: "AmazonEC2", Region: "us-east-1", Family: "Storage", Attrs: map[string]string{"volumeApiName": "gp3"}}, "0.08"},
		{pricing.Query{Service: "AmazonEC2", Region: "us-east-1", Family: "NAT Gateway", Attrs: map[string]string{"usagetype": "NatGateway-Hours"}}, "0.045"},
		{pricing.Query{Service: "AmazonS3", Region: "us-east-1", Family: "Storage", Attrs: map[string]string{"usagetype": "TimedStorage-ByteHrs"}}, "0.023"},
		{pricing.Query{Service: "AWSLambda", Region: "us-east-1", Family: "Serverless", Attrs: map[string]string{"group": "AWS-Lambda-Requests"}}, "0.0000002"},
		{pricing.Query{Service: "AWSELB", Region: "us-east-1", Family: "Load Balancer-Application", Attrs: map[string]string{"usagetype": "LoadBalancerUsage"}}, "0.0225"},
		{pricing.Query{Service: "AmazonRDS", Region: "us-east-1", Family: "Database Storage", Attrs: map[string]string{"usagetype": "RDS:GP3-Storage"}}, "0.115"},
	}
	for _, c := range checks {
		p, err := snap.Lookup(c.q)
		if err != nil {
			t.Errorf("%s: %v", c.q, err)
			continue
		}
		if got := p.UnitPrice().String(); got != c.want {
			t.Errorf("%s = %s, want %s", c.q, got, c.want)
		}
	}
}
