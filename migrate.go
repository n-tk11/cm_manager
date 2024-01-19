package main

import (
	"time"

	"go.uber.org/zap"
)

func migrateService(src string, dest string, service Service, copt CheckpointOptions, ropt RunOptions, sopt StartOptions, stopSrc bool) error {
	logger.Debug("Migrating service", zap.String("service", service.Name))
	migrateStart := time.Now()
	sErr := startServiceContainer(workers[dest], sopt)
	if sErr != nil {
		logger.Error("Error starting service's container at destination", zap.String("serviceName", service.Name), zap.String("dest", dest), zap.Error(sErr))
		return sErr
	}
	var cErr error
	ropt.ImageURL, cErr = checkpointService(src, service, copt)
	//Let user manage what port there want to use
	if cErr != nil {
		logger.Error("Error checkpoint service at source", zap.String("serviceName", service.Name), zap.String("src", src), zap.Error(cErr))

		return cErr
	}
	//startServiceContainer(workers[dest], sopt)
	//time.Sleep(200 * time.Millisecond) //If too fast ffd may not ready
	runCount = 0
	rErr := runService(workers[dest], service, ropt)
	if rErr != nil {
		logger.Error("Failed to run service on destination, will start the service on source again", zap.String("serviceName", service.Name), zap.String("src", src), zap.String("dest", dest), zap.Error(rErr))

		rErr := runService(workers[src], service, ropt)
		if rErr != nil {
			logger.Error("Failed to rerun service on source", zap.String("serviceName", service.Name), zap.String("src", src), zap.Error(rErr))
		}
		return rErr
	}
	if stopSrc {
		stErr := stopService(workers[src], service)
		if stErr != nil {
			logger.Error("Failed to stop service on source", zap.String("serviceName", service.Name), zap.String("src", src), zap.Error(rErr))
			return stErr
		}
	}
	migrateEnd := time.Since(migrateStart)
	logger.Info("Migrate service successfully", zap.String("service", service.Name), zap.String("src", src), zap.String("dest", dest), zap.Duration("time", migrateEnd))

	return nil
}
