module github.com/ONSdigital/dp-static-file-publisher

go 1.21

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

replace github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20210802184156-9742bd7fca1c+incompatible

// to avoid the following vulnerabilities:
//     - CVE-2022-29153 # pkg:golang/github.com/hashicorp/consul/api@v1.1.0
//     - sonatype-2021-1401 # pkg:golang/github.com/miekg/dns@v1.0.14
replace github.com/spf13/cobra => github.com/spf13/cobra v1.4.0

// to avoid 'sonatype-2021-4899' non-CVE Vulnerability
exclude github.com/gorilla/sessions v1.2.1

// to avoid CVE-2022-21698
exclude github.com/prometheus/client_golang v0.9.2

// to avoid 'sonatype-2020-1055' non-CVE Vulnerability
exclude github.com/go-ldap/ldap/v3 v3.1.3

// for linter warning in CI
replace github.com/blizzy78/varnamelen => github.com/blizzy78/varnamelen v0.8.0

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0
	github.com/ONSdigital/dp-component-test v0.6.5
	github.com/ONSdigital/dp-healthcheck v1.6.1
	github.com/ONSdigital/dp-kafka/v2 v2.5.0
	github.com/ONSdigital/dp-kafka/v3 v3.6.0
	github.com/ONSdigital/dp-net v1.5.0
	github.com/ONSdigital/dp-net/v2 v2.11.0
	github.com/ONSdigital/dp-s3 v1.6.0
	github.com/ONSdigital/dp-s3/v2 v2.0.0-beta.3
	github.com/ONSdigital/dp-vault v1.2.0
	github.com/ONSdigital/go-ns v0.0.0-20210410105122-6d6a140e952e
	github.com/ONSdigital/log.go/v2 v2.4.1
	github.com/aws/aws-sdk-go v1.44.76
	github.com/cucumber/godog v0.12.4
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.8.4
)

require github.com/ONSdigital/dp-api-clients-go/v2 v2.255.0

require (
	github.com/ONSdigital/dp-mongodb-in-memory v1.7.0 // indirect
	github.com/ONSdigital/s3crypto v0.0.0-20180725145621-f8943119a487 // indirect
	github.com/Shopify/sarama v1.38.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/chromedp/cdproto v0.0.0-20240116100315-4a0ec5e4c400 // indirect
	github.com/chromedp/chromedp v0.9.3 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/cucumber/messages-go/v16 v16.0.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/eapache/go-resiliency v1.5.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.8 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.6 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/vault/api v1.10.0 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/maxcnunes/httpfake v1.2.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/smarty/assertions v1.15.1 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.13.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
