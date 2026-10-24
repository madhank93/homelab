package grafana


type GrafanaSidecarDashboardsProvider struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	AllowUiUpdates *bool `field:"optional" json:"allowUiUpdates" yaml:"allowUiUpdates"`
	DisableDelete *bool `field:"optional" json:"disableDelete" yaml:"disableDelete"`
	Folder *string `field:"optional" json:"folder" yaml:"folder"`
	FoldersFromFilesStructure *bool `field:"optional" json:"foldersFromFilesStructure" yaml:"foldersFromFilesStructure"`
	FolderUid *string `field:"optional" json:"folderUid" yaml:"folderUid"`
	Name *string `field:"optional" json:"name" yaml:"name"`
	Orgid interface{} `field:"optional" json:"orgid" yaml:"orgid"`
	Type *string `field:"optional" json:"type" yaml:"type"`
}

