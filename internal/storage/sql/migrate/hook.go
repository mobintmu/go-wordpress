package migrate

import "go-wordpress/internal/config"

func RunMigrations(runner *Runner, cfg *config.Config) {
	if cfg.IsTest() {
		return
	}
	runner.Run()
}

func RunSeeder(seeder *Seeder, cfg *config.Config) {
	if cfg.IsTest() {
		return
	}
	if err := seeder.SeederRun(); err != nil {
		panic(err)
	}
}
