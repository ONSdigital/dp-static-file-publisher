module github.com/ONSdigital/dp-static-file-publisher

go 1.22

replace (
	// for linter warning in CI
	github.com/blizzy78/varnamelen => github.com/blizzy78/varnamelen v0.8.0
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go v3.2.1-0.20210802184156-9742bd7fca1c+incompatible
	// to avoid the following vulnerabilities:
	//     - CVE-2022-29153 # pkg:golang/github.com/hashicorp/consul/api@v1.1.0
	//     - sonatype-2021-1401 # pkg:golang/github.com/miekg/dns@v1.0.14
	github.com/spf13/cobra => github.com/spf13/cobra v1.4.0
	// To avoid CVE-2023-32731
	google.golang.org/grpc => google.golang.org/grpc v1.58.3
	// To fix: [CVE-2024-24786] CWE-835: Loop with Unreachable Exit Condition ('Infinite Loop')
	google.golang.org/protobuf => google.golang.org/protobuf v1.33.0
)

exclude (
	// to avoid 'sonatype-2020-1055' non-CVE Vulnerability
	github.com/go-ldap/ldap/v3 v3.1.3
	// to avoid 'sonatype-2021-4899' non-CVE Vulnerability
	github.com/gorilla/sessions v1.2.1
	// to avoid CVE-2022-21698
	github.com/prometheus/client_golang v0.9.2
)

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0
	github.com/ONSdigital/dp-api-clients-go/v2 v2.259.0
	github.com/ONSdigital/dp-component-test v0.6.5
	github.com/ONSdigital/dp-healthcheck v1.6.1
	github.com/ONSdigital/dp-kafka/v2 v2.5.0
	github.com/ONSdigital/dp-kafka/v3 v3.3.1
	github.com/ONSdigital/dp-net v1.5.0
	github.com/ONSdigital/dp-net/v2 v2.11.0
	github.com/ONSdigital/dp-s3 v1.6.0
	github.com/ONSdigital/dp-s3/v2 v2.0.0-beta.3
	github.com/ONSdigital/go-ns v0.0.0-20210410105122-6d6a140e952e
	github.com/ONSdigital/log.go/v2 v2.4.1
	github.com/aws/aws-sdk-go v1.44.76
	github.com/cucumber/godog v0.12.4
	github.com/gorilla/mux v1.8.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.8.3
)

require (
	github.com/ONSdigital/dp-mongodb-in-memory v1.2.0 // indirect
	github.com/ONSdigital/s3crypto v0.0.0-20180725145621-f8943119a487 // indirect
	github.com/Shopify/sarama v1.30.1 // indirect
	github.com/chromedp/cdproto v0.0.0-20211126220118-81fa0469ad77 // indirect
	github.com/chromedp/chromedp v0.7.6 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/cucumber/messages-go/v16 v16.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/schema v1.4.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-memdb v1.3.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/maxcnunes/httpfake v1.2.4 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/smarty/assertions v1.15.1 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.8.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
