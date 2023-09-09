package helm

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"strings"
)

const (
	caVolumeName         = "root-ca"
	defaultConfigMapName = "ca-roots"
	defaultMountPath     = "/etc/ssl/certs"
	defaultFileName      = "ca-certificates.crt"
)

func (b *UniversalOrchestratorHelmValueBuilder) caChainConfiguration() error {
	// First, determine if the values.yaml already configured the CA chain
	// If one is installed, there will be a volume called root-ca and a volume mount called root-ca

	configuredConfigMapName := ""
	mountPath := ""
	fileName := ""

	volumeExists := false
	configMapConfigured := false
	for _, volume := range b.values.Volumes {
		if volume.Name == caVolumeName {
			configuredConfigMapName = volume.ConfigMap.Name
			for _, item := range volume.ConfigMap.Items {
				fileName = item.Path
				configMapConfigured = true
			}
			volumeExists = true
		}
	}

	volumeMountExists := false
	for _, mount := range b.values.VolumeMounts {
		if mount.Name == caVolumeName {
			mountPath = mount.MountPath
			volumeMountExists = true
		}
	}

	mountPath = strings.Replace(mountPath, fmt.Sprintf("/%s", fileName), "", -1)

	// There are three copies of these variables so that changes can be accurately tracked and displayed to user
	defaultConfigMapNameSetting := ""
	if configuredConfigMapName == "" {
		defaultConfigMapNameSetting = defaultConfigMapName
	} else {
		defaultConfigMapNameSetting = configuredConfigMapName
	}

	defaultMountPathSetting := ""
	if mountPath == "" {
		defaultMountPathSetting = defaultMountPath
	} else {
		defaultMountPathSetting = mountPath
	}

	defaultFileNameSetting := ""
	if fileName == "" {
		defaultFileNameSetting = defaultFileName
	} else {
		defaultFileNameSetting = fileName
	}

	newConfigMapName := ""
	prompt := survey.Input{
		Message: "Enter the name of the configmap containing the CA chain",
		Default: defaultConfigMapNameSetting,
	}
	err := survey.AskOne(&prompt, &newConfigMapName, survey.WithValidator(survey.Required))
	if err != nil {
		return err
	}

	newFileName := ""
	prompt = survey.Input{
		Message: "Enter the file name of the certificate inside the configmap. In most cases this should be ca-certificates.crt",
		Default: defaultFileNameSetting,
	}
	err = survey.AskOne(&prompt, &newFileName, survey.WithValidator(survey.Required))
	if err != nil {
		return err
	}

	newMountPath := ""
	prompt = survey.Input{
		Message: "Enter the mount path where the certificate inside the configmap will be mounted",
		Default: defaultMountPathSetting,
	}
	err = survey.AskOne(&prompt, &newMountPath, survey.WithValidator(survey.Required))
	if err != nil {
		return err
	}

	// Print the results for the user to confirm
	printConfirmation(configuredConfigMapName, newConfigMapName, mountPath, newMountPath, fileName, newFileName, volumeExists, volumeMountExists, configMapConfigured)

	// Confirm that the user wants to save the changes
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: "Save changes?",
		Default: true,
	}
	err = survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		return err
	}

	if confirm {
		b.syncVolumeConfiguration(newConfigMapName, newMountPath, newFileName)
	}

	return b.MainMenu()
}

func (b *UniversalOrchestratorHelmValueBuilder) syncVolumeConfiguration(configMapName, mountPath, fileName string) {
	// First handle the volume/configmap
	volumeExists := false
	for i, volume := range b.values.Volumes {
		if volume.Name == caVolumeName {
			volumeExists = true
			b.values.Volumes[i].ConfigMap.Name = configMapName
			b.values.Volumes[i].ConfigMap.Items = make([]ConfigMapItem, 0)
			b.values.Volumes[i].ConfigMap.Items = append(b.values.Volumes[i].ConfigMap.Items, ConfigMapItem{Key: fileName, Path: fileName})
		}
	}

	// If the volume doesn't exist, create it
	if !volumeExists {
		newVolume := Volume{}
		newVolume.Name = caVolumeName
		newVolume.ConfigMap.Name = configMapName
		newVolume.ConfigMap.Items = append(make([]ConfigMapItem, 0), ConfigMapItem{Key: fileName, Path: fileName})
		b.values.Volumes = append(b.values.Volumes, newVolume)
	}

	// Next, handle the volume mount
	volumeMountExists := false
	for i, mount := range b.values.VolumeMounts {
		if mount.Name == caVolumeName {
			volumeMountExists = true
			b.values.VolumeMounts[i].MountPath = mountPath
			b.values.VolumeMounts[i].SubPath = fileName
		}
	}

	// If the volume mount doesn't exist, create it
	if !volumeMountExists {
		newVolumeMount := VolumeMount{}
		newVolumeMount.Name = caVolumeName
		newVolumeMount.MountPath = fmt.Sprintf("%s/%s", mountPath, fileName)
		newVolumeMount.SubPath = fileName
		b.values.VolumeMounts = append(b.values.VolumeMounts, newVolumeMount)
	}
}

func printConfirmation(oldConfigMapName, newConfigMapName, oldMountPath, newMountPath, oldFileName, newFileName string, volumeExists, volumeMountExists, configMapConfigured bool) {
	const (
		red    = "31"
		green  = "32"
		yellow = "33"
	)

	fmt.Println("The following changes will be made:")

	// Helper function to print messages with color
	printWithColor := func(prepend, text, colorCode string) {
		fmt.Printf("\033[%sm%s\033[0m%s\n", colorCode, prepend, text)
	}

	// First, print the volume configuration
	if volumeExists {
		printWithColor("~ ", "volumes:", yellow)
		printWithColor("~ ", "  - name: root-ca", yellow)
	} else {
		printWithColor("+ ", "volumes:", green)
		printWithColor("+ ", "  - name: root-ca", green)
	}

	// Next, print the configmap configuration
	if configMapConfigured {
		printWithColor("~ ", "    configMap:", yellow)
		if oldConfigMapName != newConfigMapName {
			printWithColor("+ ", fmt.Sprintf("      name: %q -> %q", oldConfigMapName, newConfigMapName), green)
		} else {
			printWithColor("~ ", fmt.Sprintf("      name: %q", newConfigMapName), yellow)
		}
	} else {
		printWithColor("+ ", "    configMap:", green)
		printWithColor("+ ", fmt.Sprintf("      name: %q", newConfigMapName), green)
	}

	// Next, print the fileName configuration
	if oldFileName != newFileName {
		printWithColor("+ ", fmt.Sprintf("      items:"), green)
		printWithColor("+ ", fmt.Sprintf("        - key: %q -> %q", oldFileName, newFileName), green)
		printWithColor("+ ", fmt.Sprintf("          path: %q -> %q", oldFileName, newFileName), green)
	} else {
		printWithColor("~ ", fmt.Sprintf("      items:"), yellow)
		printWithColor("~ ", fmt.Sprintf("        - key: %q", newFileName), yellow)
		printWithColor("~ ", fmt.Sprintf("          path: %q", newFileName), yellow)
	}

	// Finally, print the mountPath configuration
	oldMountPath = fmt.Sprintf("%s/%s", oldMountPath, oldFileName)
	newMountPath = fmt.Sprintf("%s/%s", newMountPath, newFileName)
	if volumeMountExists {
		printWithColor("~ ", "volumeMounts:", yellow)
		printWithColor("~ ", "  - name: root-ca", yellow)
		if oldMountPath != newMountPath {
			printWithColor("+ ", fmt.Sprintf("  - mountPath: %q -> %q", oldMountPath, newMountPath), green)
		} else {
			printWithColor("~ ", fmt.Sprintf("  - mountPath: %q", newMountPath), yellow)
		}
	} else {
		printWithColor("+ ", "volumeMounts:", green)
		printWithColor("+ ", "  - name: root-ca", green)
		printWithColor("+ ", fmt.Sprintf("  - mountPath: %q", newMountPath), green)
		printWithColor("+ ", fmt.Sprintf("    subPath: %q", newFileName), green)
	}
}
