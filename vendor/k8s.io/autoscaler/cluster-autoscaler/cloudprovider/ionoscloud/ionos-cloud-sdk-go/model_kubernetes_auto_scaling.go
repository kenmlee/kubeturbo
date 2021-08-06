/*
 * CLOUD API
 *
 * An enterprise-grade Infrastructure is provided as a Service (IaaS) solution that can be managed through a browser-based \"Data Center Designer\" (DCD) tool or via an easy to use API.   The API allows you to perform a variety of management tasks such as spinning up additional servers, adding volumes, adjusting networking, and so forth. It is designed to allow users to leverage the same power and flexibility found within the DCD visual tool. Both tools are consistent with their concepts and lend well to making the experience smooth and intuitive.
 *
 * API version: 5.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package ionossdk

import (
	"encoding/json"
)

// KubernetesAutoScaling struct for KubernetesAutoScaling
type KubernetesAutoScaling struct {
	// The minimum number of worker nodes that the managed node group can scale in. Should be set together with 'maxNodeCount'. Value for this attribute must be greater than equal to 1 and less than equal to maxNodeCount.
	MinNodeCount *int32 `json:"minNodeCount,omitempty"`
	// The maximum number of worker nodes that the managed node pool can scale-out. Should be set together with 'minNodeCount'. Value for this attribute must be greater than equal to 1 and minNodeCount.
	MaxNodeCount *int32 `json:"maxNodeCount,omitempty"`
}



// GetMinNodeCount returns the MinNodeCount field value
// If the value is explicit nil, the zero value for int32 will be returned
func (o *KubernetesAutoScaling) GetMinNodeCount() *int32 {
	if o == nil {
		return nil
	}

	return o.MinNodeCount
}

// GetMinNodeCountOk returns a tuple with the MinNodeCount field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *KubernetesAutoScaling) GetMinNodeCountOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.MinNodeCount, true
}

// SetMinNodeCount sets field value
func (o *KubernetesAutoScaling) SetMinNodeCount(v int32) {
	o.MinNodeCount = &v
}

// HasMinNodeCount returns a boolean if a field has been set.
func (o *KubernetesAutoScaling) HasMinNodeCount() bool {
	if o != nil && o.MinNodeCount != nil {
		return true
	}

	return false
}



// GetMaxNodeCount returns the MaxNodeCount field value
// If the value is explicit nil, the zero value for int32 will be returned
func (o *KubernetesAutoScaling) GetMaxNodeCount() *int32 {
	if o == nil {
		return nil
	}

	return o.MaxNodeCount
}

// GetMaxNodeCountOk returns a tuple with the MaxNodeCount field value
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *KubernetesAutoScaling) GetMaxNodeCountOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.MaxNodeCount, true
}

// SetMaxNodeCount sets field value
func (o *KubernetesAutoScaling) SetMaxNodeCount(v int32) {
	o.MaxNodeCount = &v
}

// HasMaxNodeCount returns a boolean if a field has been set.
func (o *KubernetesAutoScaling) HasMaxNodeCount() bool {
	if o != nil && o.MaxNodeCount != nil {
		return true
	}

	return false
}


func (o KubernetesAutoScaling) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}

	if o.MinNodeCount != nil {
		toSerialize["minNodeCount"] = o.MinNodeCount
	}
	

	if o.MaxNodeCount != nil {
		toSerialize["maxNodeCount"] = o.MaxNodeCount
	}
	
	return json.Marshal(toSerialize)
}

type NullableKubernetesAutoScaling struct {
	value *KubernetesAutoScaling
	isSet bool
}

func (v NullableKubernetesAutoScaling) Get() *KubernetesAutoScaling {
	return v.value
}

func (v *NullableKubernetesAutoScaling) Set(val *KubernetesAutoScaling) {
	v.value = val
	v.isSet = true
}

func (v NullableKubernetesAutoScaling) IsSet() bool {
	return v.isSet
}

func (v *NullableKubernetesAutoScaling) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableKubernetesAutoScaling(val *KubernetesAutoScaling) *NullableKubernetesAutoScaling {
	return &NullableKubernetesAutoScaling{value: val, isSet: true}
}

func (v NullableKubernetesAutoScaling) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableKubernetesAutoScaling) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


