package golog

// func BenchmarkGoLog(b *testing.B) {
// 	// defer Sync()
// 	InitLogger("log/golog.log", 10<<10, false)
// 	for i := 0; i < b.N; i++ {
// 		Info("msg:", "hello")
// 	}
// 	// logger := NewLog(tmpfile.Name(), 0, false)
// 	// for i := 0; i < b.N; i++ {
// 	// 	logger.String("msg", "hello").String("key1", "value1").String("key2", "value2").Log()
// 	// }
// }

// // NOTE: This does not produce a proper LTSV log since a log record does not have the "time: label.
// // This is used just for benchmark comparison.
// func BenchmarkStandardLog(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	logger := log.New(tmpfile, "", log.LstdFlags|log.Lmicroseconds)
// 	for i := 0; i < b.N; i++ {
// 		logger.Printf("level:Info\tmsg:sample log message\tkey1:%s\tkey2:%s", "value1", "value2")
// 	}
// }

// func init() {
// 	err := ltsv.RegisterLTSVEncoder()
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func BenchmarkZapJSONProductionLog(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	cfg := zap.NewProductionConfig()
// 	cfg.OutputPaths = []string{tmpfile.Name()}
// 	logger, err := cfg.Build()
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		logger.Info("sample log message", zap.String("key1", "value1"), zap.String("key2", "value2"))
// 	}
// }

// func BenchmarkZapJSONDevelopmentLog(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	cfg := zap.NewDevelopmentConfig()
// 	cfg.OutputPaths = []string{tmpfile.Name()}
// 	logger, err := cfg.Build()
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		logger.Info("sample log message", zap.String("key1", "value1"), zap.String("key2", "value2"))
// 	}
// }

// func BenchmarkZapLTSVProductionLog(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	cfg := ltsv.NewProductionConfig()
// 	cfg.OutputPaths = []string{tmpfile.Name()}
// 	logger, err := cfg.Build()
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		logger.Info("sample log message", zap.String("key1", "value1"), zap.String("key2", "value2"))
// 	}
// }

// func BenchmarkZapLTSVDevelopmentLog(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	cfg := ltsv.NewDevelopmentConfig()
// 	cfg.OutputPaths = []string{tmpfile.Name()}
// 	logger, err := cfg.Build()
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		logger.Info("sample log message", zap.String("key1", "value1"), zap.String("key2", "value2"))
// 	}
// }

// func BenchmarkZerologTimestamp(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	zerolog.TimeFieldFormat = ""
// 	logger := zerolog.New(tmpfile).With().Timestamp().Logger()
// 	for i := 0; i < b.N; i++ {
// 		logger.Info().
// 			Str("key1", "value1").
// 			Str("key2", "value2").
// 			Msg("sample log message")
// 	}
// }

// func BenchmarkZerologTimeSecondsFromEpoch(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	zerolog.TimeFieldFormat = ""
// 	logger := zerolog.New(tmpfile).With().Time("time", time.Now()).Logger()
// 	for i := 0; i < b.N; i++ {
// 		logger.Info().
// 			Str("key1", "value1").
// 			Str("key2", "value2").
// 			Msg("sample log message")
// 	}
// }

// func BenchmarkZerologRFC3339Time(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	zerolog.TimeFieldFormat = time.RFC3339
// 	logger := zerolog.New(tmpfile).With().Time("time", time.Now()).Logger()
// 	for i := 0; i < b.N; i++ {
// 		logger.Info().
// 			Str("key1", "value1").
// 			Str("key2", "value2").
// 			Msg("sample log message")
// 	}
// }

// func BenchmarkZerologRFC3339NanoTime(b *testing.B) {
// 	tmpfile, err := ioutil.TempFile("", "benchmark")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	zerolog.TimeFieldFormat = time.RFC3339Nano
// 	logger := zerolog.New(tmpfile).With().Time("time", time.Now()).Logger()
// 	for i := 0; i < b.N; i++ {
// 		logger.Info().
// 			Str("key1", "value1").
// 			Str("key2", "value2").
// 			Msg("sample log message")
// 	}
// }
