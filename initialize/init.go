package initialize


import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
)

var (
	MsqlDb       *gorm.DB             //数据库客户端
	RedisClient *redis.Client   //redis单机客户端
	//IMRedisCluster *redis.ClusterClient //redis客户端
	Config          *viper.Viper      //全局配置

)

//	提供系统初始化，全局变量
func Init(config *viper.Viper) {

	Config = config
	var err error
	//mysql配置
	MsqlDb, err = gorm.Open("mysql", config.GetString("Mysql.user")+":"+config.GetString("Mysql.password")+"@tcp("+config.GetString("Mysql.host")+":"+config.GetString("Mysql.Port")+")/"+config.GetString("Mysql.database")+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	MsqlDb.DB().SetMaxIdleConns(10)
	MsqlDb.DB().SetMaxOpenConns(100)
	// 激活链接
	if err = MsqlDb.DB().Ping(); err != nil {
		panic(err)
	}

	//	Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetString("Redis.default.addr"),
		Password: config.GetString("Redis.default.password"), // no password set
		DB:       config.GetInt("Redis.default.db"),  // use default DB
	})
	err = RedisClient.Ping().Err()
	if err != nil {
		panic(err)
	}

}