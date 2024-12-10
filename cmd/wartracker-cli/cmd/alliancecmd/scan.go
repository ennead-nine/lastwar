/*
Copyright © 2024 P4K Ennead  <ennead.tbc@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package alliancecmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
	"time"
	"wartracker/pkg/alliance"
	"wartracker/pkg/scanner"

	"github.com/spf13/cobra"
)

var imageFile string
var outputFile string
var server int64

// GetAllianceTagImage gets the alliance tag text from an alliance screenshot
func GetAllianceTagText(img image.Image) (string, error) {
	// Alliance tag rect
	px := 157
	py := 292
	rx := 48
	ry := 20

	img = scanner.GetImageRect(px, py, rx, ry, img)
	img, err := scanner.PreProcessImage(img, false, false, false)
	if err != nil {
		return "", err
	}

	return scanner.GetImageText(img, "<>0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

// GetAllianceTagImage gets the alliance tag text from an alliance screenshot
func GetAllianceNameText(img image.Image) (string, error) {
	// Alliance tag rect
	px := 204
	py := 290
	rx := 160
	ry := 24

	img = scanner.GetImageRect(px, py, rx, ry, img)
	img, err := scanner.PreProcessImage(img, false, false, false)
	if err != nil {
		return "", err
	}

	outf, _ := os.Create(fmt.Sprintf("%s-name.png", imageFile))
	defer outf.Close()
	err = png.Encode(outf, img)
	if err != nil {
		return "", err
	}

	return scanner.GetImageText(img)
}

// GetAlliancePowerText gets the alliance power text from an alliance screenshot
func GetAlliancePowerText(img image.Image) (int, error) {
	// Alliance tag rect
	px := 280
	py := 317
	rx := 96
	ry := 18

	img = scanner.GetImageRect(px, py, rx, ry, img)
	img, err := scanner.PreProcessImage(img, true, true, true)
	if err != nil {
		return 0, err
	}

	tpower, err := scanner.GetImageText(img, "0123456789")
	if err != nil {
		return 0, err
	}

	power, err := strconv.Atoi(tpower)
	if err != nil {
		return 0, err
	}

	return power, nil
}

// GetAllianceGiftImage gets the alliance gift level text from an alliance screenshot
func GetAllianceGiftText(img image.Image) (int, error) {
	// Alliance tag rect
	px := 356
	py := 351
	rx := 19
	ry := 15

	img = scanner.GetImageRect(px, py, rx, ry, img)
	img, err := scanner.PreProcessImage(img, true, true, true)
	if err != nil {
		return 0, err
	}

	tgift, err := scanner.GetImageText(img, "0123456789")
	if err != nil {
		return 0, err
	}

	gift, err := strconv.Atoi(tgift)
	if err != nil {
		return 0, err
	}

	return gift, nil
}

// GetAllianceGiftImage gets the alliance gift level text from an alliance screenshot
func GetAllianceMemberText(img image.Image) (int, error) {
	// Alliance tag rect
	px := 316
	py := 366
	rx := 28
	ry := 16

	img = scanner.GetImageRect(px, py, rx, ry, img)
	img, err := scanner.PreProcessImage(img, true, true, true)
	if err != nil {
		return 0, err
	}

	tmemcount, err := scanner.GetImageText(img, "0123456789")
	if err != nil {
		return 0, err
	}

	memcount, err := strconv.Atoi(tmemcount)
	if err != nil {
		return 0, err
	}

	return memcount, nil
}

// ScanAlliance pre-processes the given image file and scans it with tessaract
// into an alliance.Alliance struct
func ScanAlliance() (*alliance.Alliance, error) {
	var a alliance.Alliance
	var d alliance.Data

	img, err := scanner.SetImageDensity(inputFile, 300)
	if err != nil {
		return nil, err
	}

	// Setup alliance
	tag, err := GetAllianceTagText(img)
	if err != nil {
		return nil, err
	}
	name, err := GetAllianceNameText(img)
	if err != nil {
		return nil, err
	}
	power, err := GetAlliancePowerText(img)
	if err != nil {
		return nil, err
	}
	gift, err := GetAllianceGiftText(img)
	if err != nil {
		return nil, err
	}
	memcount, err := GetAllianceMemberText(img)
	if err != nil {
		return nil, err
	}

	d.Date = time.Now().Format(time.DateOnly)
	d.Tag = tag
	d.Name = name
	d.Power = int64(power)
	d.GiftLevel = int64(gift)
	d.MemberCount = int64(memcount)
	a.Data = append(a.Data, d)
	a.Server = server

	err = a.GetByTag(d.Tag)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		fmt.Printf("A new alliance will need to be created from this data.  Please run 'wartracker-cli alliance new -o %s' after verifying the data\n", outputFile)
	} else {
		fmt.Printf("This alliance already exists. To add the new data run 'wartracker-cli alliance add -o %s' to add the new data.\n", outputFile)
	}

	a.Data = a.Data[:1]

	j, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(outputFile, j, 0644)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans an alliance screenshot into a json file.",
	Long: `Scan takes an alliance screenshot and Marshals an alliance object 
	into json for cleanup.  Running wartracjer-cli alliance create with the 
	cleaned json will create an alliance object in the database.
	
	Example: wartracker-cli alliance scan -i alliance.png`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := ScanAlliance()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	allianceCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// canmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	scanCmd.Flags().StringVarP(&imageFile, "image", "i", "", "image file (PNG) to scan for alliance data")
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "JSON file to output alliance data to")
	scanCmd.Flags().Int64VarP(&server, "server", "s", 1, "Alliance's server number")
}
