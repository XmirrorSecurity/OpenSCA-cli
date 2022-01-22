/*
 * @Description: pargress bar
 * @Date: 2022-01-20 15:48:46
 */

package bar

import "github.com/schollz/progressbar/v3"

var (
	Dir        = progressbar.Default(-1, "scan dir")
	Archive    = progressbar.Default(-1, "unarchive")
	Maven      = progressbar.Default(-1, "parse maven indirect dependency")
	Npm        = progressbar.Default(-1, "parse npm indirect dependency")
	Composer   = progressbar.Default(-1, "parse composer indirect dependency")
	Dependency = progressbar.Default(-1, "parse project dependency")
)
