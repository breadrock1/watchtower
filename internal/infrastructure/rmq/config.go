package rmq

type Config struct {
	Address    string `mapstructure:"address"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
	QueueName  string `mapstructure:"queue_name"`
}
