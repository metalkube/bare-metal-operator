/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Additional data describing the firmware setting
type SettingSchema struct {

	// The type of setting.
	// +kubebuilder:validation:Enum=Enumeration;String;Integer;Boolean;Password
	AttributeType string `json:"attribute_type,omitempty"`

	// The allowable value for an Enumeration type setting.
	AllowableValues []string `json:"allowable_values,omitempty"`

	// The lowest value for an Integer type setting.
	LowerBound *int `json:"lower_bound,omitempty"`

	// The highest value for an Integer type setting.
	UpperBound *int `json:"upper_bound,omitempty"`

	// Minimum length for a String type setting.
	MinLength *int `json:"min_length,omitempty"`

	// Maximum length for a String type setting.
	MaxLength *int `json:"max_length,omitempty"`

	// Whether or not this setting is read only.
	ReadOnly *bool `json:"read_only,omitempty"`

	// Whether or not a reset is required after changing this setting.
	ResetRequired *bool `json:"reset_required,omitempty"`

	// Whether or not this setting's value is unique to this node, e.g.
	// a serial number.
	Unique *bool `json:"unique,omitempty"`
}

// FirmwareSchemaSpec defines the desired state of FirmwareSchema
type FirmwareSchemaSpec struct {

	// The hardware vendor associated with this schema
	// +optional
	HardwareVendor string `json:"hardwareVendor,omitempty"`

	// The hardware model associated with this schema
	// +optional
	HardwareModel string `json:"hardwareModel,omitempty"`

	// Map of firmware name to schema
	Schema map[string]SettingSchema `json:"schema" required:"true"`
}

//+kubebuilder:object:root=true

// FirmwareSchema is the Schema for the firmwareschemas API
type FirmwareSchema struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FirmwareSchemaSpec `json:"spec,omitempty"`
}

// Check whether the setting's name and value is valid using the schema
func (host *FirmwareSchema) CheckSettingIsValid(name string, value intstr.IntOrString, schemas map[string]SettingSchema) error {

	schema, ok := schemas[name]
	if !ok {
		return fmt.Errorf("Setting %s is not in the schema", name)
	}

	if schema.ReadOnly != nil && *schema.ReadOnly == true {
		return fmt.Errorf("Setting %s is ReadOnly", name)
	}

	// Check if valid based on type
	switch schema.AttributeType {
	case "Enumeration":
		for _, av := range schema.AllowableValues {
			if value.String() == av {
				return nil
			}
		}
		return fmt.Errorf("Setting %s uses invalid enum %s", name, value.String())

	case "Integer":
		if schema.LowerBound == nil || schema.UpperBound == nil {
			// return true if no settings to check validity
			return nil
		}
		if value.IntValue() >= *schema.LowerBound && value.IntValue() <= *schema.UpperBound {
			return nil
		}
		return fmt.Errorf("Setting %s integer %s is out of range, %d - %d", name, value.String(), *schema.LowerBound, *schema.UpperBound)

	case "String":
		if schema.MinLength == nil || schema.MaxLength == nil {
			// return true if no settings to check validity
			return nil
		}
		strLen := len(value.String())
		if strLen >= *schema.MinLength && strLen <= *schema.MaxLength {
			return nil
		}
		return fmt.Errorf("Setting %s string %s length is out of range, %d - %d", name, value.String(), *schema.MinLength, *schema.MaxLength)

	case "Boolean":
		if value.String() == "true" || value.String() == "false" {
			return nil
		}
		return fmt.Errorf("Setting %s is not a boolean - %s", name, value.String())

	case "Password":
		// Prevent sets of password types
		return fmt.Errorf("Setting %s is a Password type", name)

	case "":
		// allow the set as BIOS registry fields may not have been available
		return nil

	default:
		// Unexpected attribute type
		return fmt.Errorf("Setting %s has an unexpected attribute type %s", name, schema.AttributeType)
	}
}

//+kubebuilder:object:root=true

// FirmwareSchemaList contains a list of FirmwareSchema
type FirmwareSchemaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirmwareSchema `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirmwareSchema{}, &FirmwareSchemaList{})
}
