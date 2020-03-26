package usecase

import (
	"fmt"
	"strings"

	"git.sdn.sbrf.ru/users/tihonov-id/repos/nw-pr-lb/domain"
)

func tunnelsRemove(deployedEntities map[string][]string,
	tunnelConfig domain.TunnelMaker,
	requestUUID string) error {
	errors := []error{}
	if createdTunnelFilesArray, inMap := deployedEntities["createdTunnelFiles"]; inMap {
		if err := tunnelConfig.RemoveCreatedTunnelFiles(createdTunnelFilesArray, requestUUID); err != nil {
			errors = append(errors, err)
		}
	}
	if newTunnelsArray, inMap := deployedEntities["newTunnels"]; inMap {
		if err := tunnelConfig.ExecuteCommandForTunnels(newTunnelsArray, "down", requestUUID); err != nil {
			errors = append(errors, err)
		}
	}
	return combineErrors(errors)
}

func keepalivedConfigRemove(deployedEntities map[string][]string,
	keepalivedConfig domain.KeepalivedCustomizer,
	requestUUID string) error {
	errors := []error{}
	// if rowsInKeepalivedConfig, inMap := deployedEntities["newKeepalivedDConfigFileName"]; inMap {
	// rework!!!
	// if err := keepalivedConfig.RemoveRowFromKeepalivedConfigFile(rowsInKeepalivedConfig[0], requestUUID); err != nil { // TODO: remove that
	// 	errors = append(errors, err)
	// }
	// }
	if newFullKeepalivedDConfigFilePath, inMap := deployedEntities["newFullKeepalivedDConfigFilePath"]; inMap {
		if err := keepalivedConfig.RemoveKeepalivedDConfigFile(newFullKeepalivedDConfigFilePath[0], requestUUID); err != nil {
			errors = append(errors, err)
		}
	}
	if fullPathToEnabledKeepalivedDFile, inMap := deployedEntities["fullPathToEnabledKeepalivedDFile"]; inMap {
		if err := keepalivedConfig.RemoveKeepalivedSymlink(fullPathToEnabledKeepalivedDFile[0], requestUUID); err != nil {
			errors = append(errors, err)
		}
	}
	return combineErrors(errors)
}

func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var errorsStringSlice []string
	for _, err := range errors {
		errorsStringSlice = append(errorsStringSlice, err.Error())
	}
	return fmt.Errorf(strings.Join(errorsStringSlice, "\n"))
}
