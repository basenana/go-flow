/*
   Copyright 2023 Go-Flow Authors

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

package flow

import (
	"github.com/basenana/go-flow/types"
	"github.com/basenana/go-flow/utils"
)

var logger = utils.NewLogger("go-flow")

func IsFinishedStatus(sts string) bool {
	switch sts {
	case types.SucceedStatus, types.FailedStatus, types.CanceledStatus, types.ErrorStatus:
		return true
	default:
		return false
	}
}
