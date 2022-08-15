// Copyright Aeraki Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeepCopyMap deep copy a map
func DeepCopyMap(m map[string]string) map[string]string {
	cp := make(map[string]string)
	for k, v := range m {
		cp[k] = v
	}
	return cp
}


func IsRealError(err error) bool {
	return err != nil && !errors.IsNotFound(err)
}

func IsRetryableError(err error) bool {
	return errors.IsInternalError(err) || errors.IsResourceExpired(err) || errors.IsServerTimeout(err) ||
		errors.IsServiceUnavailable(err) || errors.IsTimeout(err) || errors.IsTooManyRequests(err) ||
		errors.ReasonForError(err) == metav1.StatusReasonUnknown
}

func IsNotFound(err error) bool {
	return err != nil && errors.IsNotFound(err)
}

func Struct2JSON(ojb interface{}) interface{} {
	b, err := json.Marshal(ojb)
	if err != nil {
		return ojb
	}
	return string(b)
}
