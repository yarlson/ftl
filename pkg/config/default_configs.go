package config

// defaultConfigs holds "base name" → default configuration
// (image, ports, volumes, environment variables, container settings, etc.).
var defaultConfigs = map[string]Dependency{
	"mysql": {
		Name:    "mysql",
		Image:   "mysql:latest",
		Ports:   []int{3306},
		Volumes: []string{"mysql_data:/var/lib/mysql"},
		Env: []string{
			"MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}",
		},
	},
	"postgres": {
		Name:    "postgres",
		Image:   "postgres:latest",
		Ports:   []int{5432},
		Volumes: []string{"postgres_data:/var/lib/postgresql/data"},
		Env: []string{
			"POSTGRES_PASSWORD=${POSTGRES_PASSWORD}",
		},
	},
	"postgresql": {
		Name:    "postgres",
		Image:   "postgres:latest",
		Ports:   []int{5432},
		Volumes: []string{"postgres_data:/var/lib/postgresql/data"},
		Env: []string{
			"POSTGRES_PASSWORD=${POSTGRES_PASSWORD}",
		},
	},
	"elasticsearch": {
		Name:    "elasticsearch",
		Image:   "elasticsearch:latest",
		Ports:   []int{9200, 9300},
		Volumes: []string{"es_data:/usr/share/elasticsearch/data"},
		Env: []string{
			"discovery.type=single-node",
		},
		Container: &Container{
			ULimits: []ULimit{
				{
					Name: "nofile",
					Soft: 65535,
					Hard: 65535,
				},
			},
		},
	},
	"mongodb": {
		Name:    "mongodb",
		Image:   "mongo:latest",
		Ports:   []int{27017},
		Volumes: []string{"mongo_data:/data/db"},
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=${MONGO_ROOT_USER:mongouser}",
			"MONGO_INITDB_ROOT_PASSWORD=${MONGO_ROOT_PASSWORD}",
		},
	},
	"redis": {
		Name:    "redis",
		Image:   "redis:latest",
		Ports:   []int{6379},
		Volumes: []string{"redis_data:/data"},
	},
	"memcached": {
		Name:  "memcached",
		Image: "memcached:latest",
		Ports: []int{11211},
	},
	"rabbitmq": {
		Name:    "rabbitmq",
		Image:   "rabbitmq:latest",
		Ports:   []int{5672, 15672}, // main + management
		Volumes: []string{"rabbitmq_data:/var/lib/rabbitmq"},
		Container: &Container{
			ULimits: []ULimit{
				{
					Name: "nofile",
					Soft: 65535,
					Hard: 65535,
				},
			},
		},
	},
}
