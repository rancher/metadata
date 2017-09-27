package content

const (
	V1 = "2015-12-19"
	V2 = "2016-07-29"
	V3 = "2017-04-22"
)

var versionList = []string{
	V1,
	V2,
	V3,
	"latest",
}

var VersionMap = map[string]interface{}{}

func init() {
	for _, version := range versionList {
		VersionMap[version] = "/" + version
	}
}

func GetEnvironment(store Store, version, clientIP string) (interface{}, bool) {
	if version == "/" {
		return VersionMap, true
	}

	if version == "latest" {
		version = versionList[len(versionList)-2]
	}

	if _, ok := VersionMap[version]; !ok {
		return nil, false
	}

	env := store.Environment(Client{
		Version: version,
		IP:      clientIP,
	})

	return env, env != nil
}
