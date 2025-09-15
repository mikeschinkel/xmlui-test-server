package cfgldr

//
//func init() {
//	registerDatabaseConfig(&GenericDBConfig{})
//}
//
//var _ DatabaseConfig = (*GenericDBConfig)(nil)
//
//type GenericDBConfig struct {
//	connectString string
//	port          int
//	extensions    []DBExtensionConfig
//	sourceFile    string
//}
//
//func (g GenericDBConfig) SetSchemaQueries(strings []string) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (g GenericDBConfig) SchemaQueries() []string {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (g GenericDBConfig) OnOpenQueries() []string {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (g GenericDBConfig) SourceFile() string {
//	return g.sourceFile
//}
//
//func (g GenericDBConfig) OnLoadQueries() []string {
//	return []string{}
//}
//
//func (g GenericDBConfig) Normalize(sourceFile string) (err error) {
//	g.sourceFile = sourceFile
//	// GenericDBConfig is never really used so we don't need this method implemented
//	return err
//}
//
//func NewGenericDBConfig(args GenericDBConfigArgs) *GenericDBConfig {
//	return &GenericDBConfig{
//		connectString: args.ConnectString,
//		port:          args.Port,
//		extensions:    args.Extensions,
//	}
//}
//
//type GenericDBConfigArgs struct {
//	ConnectString string
//	Port          int
//	Extensions    []DBExtensionConfig
//}
//
//func (g GenericDBConfig) DatabaseConfig() {
//}
//
//func (g GenericDBConfig) DatabaseType() DatabaseType {
//	return GenericDatabase
//}
//
//func (g GenericDBConfig) ConnectString() string {
//	return g.connectString
//}
//
//func (g GenericDBConfig) Port() int {
//	return g.port
//}
//
//func (g GenericDBConfig) DBExtensions() []DBExtensionConfig {
//	return g.extensions
//}
