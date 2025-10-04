package dt

import (
	"strings"
)

type CategoryType string

const (
	CategoryTypeUnknown CategoryType = ""
	CategoryTypeItem    CategoryType = "item"
)

type PropertyType string

const (
	PropertyTypeUnknown     PropertyType = ""
	PropertyTypeTag         PropertyType = "Tags"
	PropertyTypeDisplayType PropertyType = "DisplayType"
	PropertyTypeHoldType    PropertyType = "HoldType"
	PropertyTypeAction0     PropertyType = "Action0"
	PropertyTypeHitSounds   PropertyType = "HitSounds"
)

type DataTool struct {
	Entries []DataEntry
}

type DataEntry struct {
	CategoryType  CategoryType
	CategoryName  string
	PropertyType  PropertyType
	PropertyValue string
	PropertyEntry []SubDataEntry
}

type SubDataEntry struct {
	PropertyType  PropertyType
	PropertyValue string
	PropertyEntry []SubDataEntry
}

// item.meleeToolRepairT0StoneAxe.Tags=axe,melee,light,tool,longShaft,repairTool,miningTool,attStrength,perkMiner69r,perkMotherLode,perkTheHuntsman,canHaveCosmetic,harvestingSkill,corpseRemoval
// item.meleeToolRepairT0StoneAxe.DisplayType=meleeRepairTool
// item.meleeToolRepairT0StoneAxe.HoldType=32

func (e DataEntry) String() string {
	var sb strings.Builder
	sb.WriteString(string(e.CategoryType))
	if e.CategoryName != "" {
		sb.WriteString(".")
		sb.WriteString(e.CategoryName)
	}
	if len(e.PropertyEntry) == 0 {
		if e.PropertyType != "" {
			sb.WriteString(".")
			sb.WriteString(string(e.PropertyType))
		}

		if e.PropertyValue != "" {
			sb.WriteString("=")
			sb.WriteString(e.PropertyValue)
		}
		return sb.String()
	}
	for _, subEntry := range e.PropertyEntry {
		var subSb strings.Builder
		subSb.WriteString(sb.String())
		if subEntry.PropertyType != "" {
			subSb.WriteString(".")
			subSb.WriteString(string(subEntry.PropertyType))
		}
		for _, subSubEntry := range subEntry.PropertyEntry {
			var subSubSb strings.Builder
			subSubSb.WriteString(subSb.String())
			if subSubEntry.PropertyType != "" {
				subSubSb.WriteString(".")
				subSubSb.WriteString(string(subSubEntry.PropertyType))
			}
			if subSubEntry.PropertyValue != "" {
				subSubSb.WriteString("=")
				subSubSb.WriteString(subSubEntry.PropertyValue)
			}
			sb.WriteString(subSubSb.String())
			sb.WriteString("\n")
		}

		if subEntry.PropertyValue == "" {
			continue
		}
		subSb.WriteString("=")
		subSb.WriteString(subEntry.PropertyValue)

		subSb.WriteString("\n")
		sb.WriteString(subSb.String())
	}
	return sb.String()
}

func (e *DataTool) String() string {
	var sb strings.Builder
	for _, entry := range e.Entries {
		sb.WriteString(entry.String())
	}
	return sb.String()
}
