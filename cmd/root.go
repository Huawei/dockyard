/*
Copyright 2015 The ContainerOps Authors All rights reserved.

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

package cmd

import (
	"github.com/spf13/cobra"
)

//RootCmd is root cmd of dockyard.
var RootCmd = &cobra.Command{
	Use:   "dockyard",
	Short: "dockyard is a container and artifact repository",
	Long: `Dockyard is a container and artifact repository storing and distributing container image, 
  software artifact and virtual images of KVM or XEN. We hosting a public service in https://dockyard.sh.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}
