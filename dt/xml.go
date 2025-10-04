package dt

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FormatType int

const (
	FormatTypeUnknown FormatType = iota
	FormatTypeItem
)

type ItemsXML struct {
	XMLName xml.Name  `xml:"items"`
	Items   []ItemXML `xml:"item"`
}

type ItemXML struct {
	Name        string        `xml:"name,attr"`
	Extends     string        `xml:"extends,attr,omitempty"`
	Description string        `xml:"description,attr,omitempty"`
	Properties  []PropertyXML `xml:"property"`
	// Add other fields as needed based on the XML structure
}

type PropertyXML struct {
	Name       string        `xml:"name,attr"`
	Class      string        `xml:"class,attr,omitempty"`
	Value      string        `xml:"value,attr"`
	Properties []PropertyXML `xml:"property"`
}

func FromXML(path string) (*DataTool, error) {

	fileName := filepath.Base(path)
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer r.Close()

	// For now, assume FormatTypeItem if file name contains "item"
	switch strings.ToLower(fileName) {
	case "items.xml":
		return FromXMLReader(FormatTypeItem, r)
	}
	return nil, fmt.Errorf("unknown file format: %s", fileName)
}

func FromXMLReader(formatType FormatType, r io.ReadSeeker) (*DataTool, error) {
	switch formatType {
	case FormatTypeItem:
		var itemsXML ItemsXML
		decoder := xml.NewDecoder(r)
		err := decoder.Decode(&itemsXML)
		if err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}
		//fmt.Printf("%+v\n", itemsXML)
		data := &DataTool{}
		err = data.DecodeXML(itemsXML)
		if err != nil {
			return nil, fmt.Errorf("from xml: %w", err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("unknown format type: %v", formatType)
}

func (dt *DataTool) DecodeXML(itemsXML ItemsXML) error {
	for _, item := range itemsXML.Items {
		//fmt.Println("Item:", item.Name)

		for _, prop := range item.Properties {

			entry := DataEntry{
				CategoryType: CategoryTypeItem,
				CategoryName: item.Name,
			}
			propType := PropertyType(prop.Name)
			if propType == PropertyTypeUnknown {
				propType = PropertyType(prop.Class)
				if propType == PropertyTypeUnknown {
					return fmt.Errorf("unknown property type: name:%s class:%s from %+v", prop.Name, prop.Class, prop)
				}
			}
			//fmt.Println("  Property:", prop.Name, "=", prop.Value, "as", propType)
			propEntry := SubDataEntry{
				PropertyType:  propType,
				PropertyValue: prop.Value,
			}
			entry.PropertyEntry = append(entry.PropertyEntry, propEntry)

			dt.Entries = append(dt.Entries, entry)
			for _, subProp := range prop.Properties {
				subPropType := PropertyType(subProp.Name)
				if subPropType == PropertyTypeUnknown {
					subPropType = PropertyType(subProp.Class)
					if subPropType == PropertyTypeUnknown {
						return fmt.Errorf("unknown sub-property type: name:%s class:%s from %+v", subProp.Name, subProp.Class, subProp)
					}
				}
				//fmt.Println("    SubProperty:", subProp.Name, "=", subProp.Value, "as", subPropType)
				subPropEntry := SubDataEntry{
					PropertyType:  subPropType,
					PropertyValue: subProp.Value,
				}
				entry.PropertyEntry[len(entry.PropertyEntry)-1].PropertyEntry = append(entry.PropertyEntry[len(entry.PropertyEntry)-1].PropertyEntry, subPropEntry)
			}

		}
	}

	return nil
}
