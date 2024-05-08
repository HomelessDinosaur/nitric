package embeds

// embed: cloudbuild-create-db.yaml
var cloudbuild_CreateDatabaseTemplate []byte

func GetCloudBuildCreateDatabaseConfig() string {
	return string(cloudbuild_CreateDatabaseTemplate)
}
