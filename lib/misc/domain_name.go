// Copyright (C) 2018 Manabu Sonoda.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package misc

import "strings"

func FQDN(dn string) string {
	bs := []byte(dn)
	if len(bs) == 0 {
		return "."
	}
	if bs[len(bs)-1] != '.' {
		return dn + "."
	}
	return dn
}

func Labels(dn string) []string {
	dn = FQDN(dn)
	labels := strings.Split(dn, ".")
	return labels[:len(labels)-1]
}

func IsDomainName(dn string) bool {
	return false
}
