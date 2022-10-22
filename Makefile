include api/.dev.vars
export

db-dump: atlas-cli
	./atlas-cli schema inspect -u "mysql://$(DATABASE_USER):$(DATABASE_PASSWORD)@eu-central.connect.psdb.cloud/shopmon?tls=true" > schema.hcl
db-apply: atlas-cli
	./atlas-cli schema apply -f schema.hcl -u "mysql://$(DATABASE_USER):$(DATABASE_PASSWORD)@eu-central.connect.psdb.cloud/shopmon?tls=true"
atlas-cli:
	curl -L https://release.ariga.io/atlas/atlas-linux-amd64-v0.7.0 -o atlas-cli
	chmod +x atlas-cli