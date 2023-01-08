/*
Copyright © 2023 Simon Emms <simon@simonemms.com>

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
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/mrsimonemms/historian/pkg/drivers"
	"github.com/spf13/cobra"
)

var mirrorOpts struct {
	Driver   drivers.Driver
	FilePath string
}

// mirrorCmd represents the mirror command
var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirror your terminal history",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// There must be one argument, which is the file we're watching
		cobra.CheckErr(cobra.ExactArgs(1)(cmd, args))

		mirrorOpts.FilePath = args[0]

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Use a persistent post run command as each of the subcommands is there
		// to create the configuration only. The execution happens here.
		//
		// There can be only one PersistentPostRun command.

		// Validate the driver
		cobra.CheckErr(mirrorOpts.Driver.Login())

		// Create new watcher.
		watcher, err := fsnotify.NewWatcher()
		cobra.CheckErr(err)

		defer func() {
			err := watcher.Close()
			cobra.CheckErr(err)
		}()

		// Start listening for events.
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					log.Println("event:", event)
					if event.Has(fsnotify.Write) {
						log.Println("modified file:", event.Name)
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()

		// Add a path.
		err = watcher.Add(mirrorOpts.FilePath)
		cobra.CheckErr(err)

		log.Println("Watching for changes in", mirrorOpts.FilePath)

		// Block main goroutine forever.
		<-make(chan struct{})
	},
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
}
