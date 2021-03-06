package initer

import (
	"auth_backend/configurator"
	"auth_backend/controller"
	"auth_backend/controller/hash"
	"auth_backend/dbInit"
	"auth_backend/controller/mailer"
	"auth_backend/errors"
	"auth_backend/logger"
	"auth_backend/postgres"
)

func InitPackages(configFileName string) *errors.Error {
	/*
	**	Read config
	 */
	print("Считываю конфигурационный файл\t\t- ")
	if err := configurator.SetConfigFile(configFileName); err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return errors.ConfigurationFail.SetArgs("Не могу считать файл "+configFileName,
			"Cant read file "+configFileName).SetOrigin(err)
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	logger
	 */
	print("Настраиваю пакет logger\t\t\t- ")
	cfgLogger := logger.GetConfig()
	if err := configurator.ParsePackageConfig(cfgLogger, "logger"); err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return errors.ConfigurationFail.SetArgs("logger", "logger").SetOrigin(err)
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	postgres
	 */
	print("Настраиваю пакет postgres\t\t- ")
	cfgPostgres := postgres.GetConfig()
	if err := configurator.ParsePackageConfig(cfgPostgres, "postgres"); err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return errors.ConfigurationFail.SetArgs("postgres", "postgres").SetOrigin(err)
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	initialization of storage
	 */
	print("Инициализирую пакет postgres\t\t- ")
	if Err := postgres.Init(); Err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return Err
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	hash
	 */
	print("Настраиваю пакет hash\t\t\t- ")
	cfgHash := hash.GetConfig()
	if err := configurator.ParsePackageConfig(cfgHash, "hash"); err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return errors.ConfigurationFail.SetArgs("controller/hash", "controller/hash").SetOrigin(err)
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	mailer
	 */
	print("Настраиваю пакет mailer\t\t\t- ")
	cfgMailer := mailer.GetConfig()
	if err := configurator.ParsePackageConfig(cfgMailer, "mailer"); err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return errors.ConfigurationFail.SetArgs("controller/mailer", "controller/mailer").SetOrigin(err)
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	mailer
	 */
	print("Настраиваю пакет controller\t\t- ")
	cfgController := controller.GetConfig()
	if err := configurator.ParsePackageConfig(cfgController, "server"); err != nil {
		println(logger.RED + "ошибка" + logger.NO_COLOR)
		return errors.ConfigurationFail.SetArgs("controller/mailer", "controller/mailer").SetOrigin(err)
	}
	println(logger.GREEN + "успешно" + logger.NO_COLOR)
	/*
	**	Recreate tables if it needed
	 */
	if cfgPostgres.RecreateTables {
		if Err := dbInit.RecreateTables(); Err != nil {
			return Err
		}
	}
	return nil
}

func CloseAllConnections() *errors.Error {
	if Err := postgres.Close(); Err != nil {
		return Err
	}
	return nil
}
