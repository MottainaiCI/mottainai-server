/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package template

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

type AvatarBuilder interface {
	GetAvatar(string) string
}

type GetAvataaarsCom struct {
	AvatarStyle     string
	Top             string
	Accessories     string
	HairColor       string
	FacialHair      string
	FacialHairColor string
	Clothes         string
	ColorFabric     string
	Eyes            string
	Eyebrow         string
	Mouth           string
	Skin            string
}

func NewGetAvataaarsCom() *GetAvataaarsCom {
	return &GetAvataaarsCom{
		AvatarStyle:     "Circle",
		Top:             "ShortHairShaggyMullet",
		Accessories:     "Prescription01",
		HairColor:       "Platinum&facialHairType",
		FacialHair:      "BeardMagestic",
		FacialHairColor: "Blonde",
		Clothes:         "Hoodie",
		ColorFabric:     "Red",
		Eyes:            "Cry",
		Eyebrow:         "FlatNatural",
		Mouth:           "Tongue",
		Skin:            "Pale'",
	}
}

func (g *GetAvataaarsCom) GenAvatarStyle(hash string) {
	m := map[int]string{
		0: "Circle",
		1: "Transparent",
	}

	s, _ := hex.DecodeString(hash)
	g.AvatarStyle = m[int(s[0])%2]
}

func (g *GetAvataaarsCom) GenAccessories(hash string) {
	m := map[int]string{
		0: "Blank",
		1: "Kurt",
		2: "Prescription01",
		3: "Prescription02",
		4: "Round",
		5: "Sunglasses",
		6: "Wayfarers",
	}

	s, _ := hex.DecodeString(hash)
	g.Accessories = m[int(s[3])%7]
}

func (g *GetAvataaarsCom) GenHairColor(hash string) {
	m := map[int]string{
		0: "Auburn",
		1: "Black",
		2: "Blonde",
		3: "BlondeGolden",
		4: "Brown",
		5: "BrownDark",
		6: "PastelPink",
		7: "Platinum",
		8: "Red",
		9: "SilverGray",
	}

	s, _ := hex.DecodeString(hash)
	g.HairColor = m[int(s[10])%10]
}

func (g *GetAvataaarsCom) GenClothes(hash string) {
	m := map[int]string{
		0: "BlazerShirt",
		1: "BlazerSweater",
		2: "CollarSweater",
		3: "GraphicShirt",
		4: "Hoodie",
		5: "Overall",
		6: "ShirtCrewNeck",
		7: "ShirtScoopNeck",
		8: "ShirtVNeck",
	}

	s, _ := hex.DecodeString(hash)
	g.Clothes = m[int(s[11])%9]
}

func (g *GetAvataaarsCom) GenColorFabric(hash string) {
	m := map[int]string{
		0:  "Black",
		1:  "Blue01",
		2:  "Blue02",
		3:  "Blue03",
		4:  "Gray01",
		5:  "Gray02",
		6:  "Heather",
		7:  "PastelBlue",
		8:  "PastelGreen",
		9:  "PastelOrange",
		10: "PastelRed",
		11: "PastelYellow",
		12: "Pink",
		13: "Red",
		14: "White",
	}

	s, _ := hex.DecodeString(hash)
	g.ColorFabric = m[int(s[9])%15]
}

func (g *GetAvataaarsCom) GenEyes(hash string) {
	m := map[int]string{
		0:  "Close",
		1:  "Cry",
		2:  "Default",
		3:  "Dizzy",
		4:  "EyeRoll",
		5:  "Happy",
		6:  "Hearts",
		7:  "Side",
		8:  "Squint",
		9:  "Surprised",
		10: "Wink",
		11: "WinkWacky",
	}

	s, _ := hex.DecodeString(hash)
	g.Eyes = m[int(s[10])%12]
}

func (g *GetAvataaarsCom) GenEyebrow(hash string) {
	m := map[int]string{
		0:  "Angry",
		1:  "AngryNatural",
		2:  "Default",
		3:  "DefaultNatural",
		4:  "FlatNatural",
		5:  "RaisedExcited",
		6:  "RaisedExcitedNatural",
		7:  "SadConcerned",
		8:  "SadConcernedNatural",
		9:  "UnibrowNatural",
		10: "UpDown",
		11: "UpDownNatural",
	}

	s, _ := hex.DecodeString(hash)
	g.Eyebrow = m[int(s[12])%12]
}

func (g *GetAvataaarsCom) GenMouth(hash string) {
	m := map[int]string{
		0:  "Concerned",
		1:  "Default",
		2:  "Disbelief",
		3:  "Eating",
		4:  "Grimace",
		5:  "Sad",
		6:  "ScreamOpen",
		7:  "Serious",
		8:  "Smile",
		9:  "Tongue",
		10: "Twinkle",
		11: "Vomit",
	}

	s, _ := hex.DecodeString(hash)
	g.Mouth = m[int(s[13])%12]
}

func (g *GetAvataaarsCom) GenSkin(hash string) {
	m := map[int]string{
		0: "Tanned",
		1: "Yellow",
		2: "Pale",
		3: "Light",
		4: "Brown",
		5: "DarkBrown",
		6: "Black",
	}

	s, _ := hex.DecodeString(hash)
	g.Skin = m[int(s[5])%7]
}

func (g *GetAvataaarsCom) GenFacialHair(hash string) {
	m := map[int]string{
		0: "Blank",
		1: "BeardMedium",
		2: "BeardLight",
		3: "BeardMagestic",
		4: "MoustacheFancy",
		5: "MoustacheMagnum",
	}

	s, _ := hex.DecodeString(hash)
	g.FacialHair = m[int(s[15])%6]
}

func (g *GetAvataaarsCom) GenFacialHairColor(hash string) {
	m := map[int]string{
		0: "Auburn",
		1: "Black",
		2: "Blonde",
		3: "BlondeGolden",
		4: "Brown",
		5: "BrownDark",
		7: "Platinum",
		8: "Red",
	}

	s, _ := hex.DecodeString(hash)
	g.FacialHairColor = m[int(s[3])%9]
}

func (g *GetAvataaarsCom) GenTop(hash string) {
	m := map[int]string{
		0:  "NoHair",
		1:  "Eyepatch",
		2:  "Hat",
		3:  "Hijab",
		4:  "Turban",
		5:  "WinterHat1",
		6:  "WinterHat2",
		7:  "WinterHat3",
		8:  "WinterHat4",
		9:  "LongHairBigHair",
		10: "LongHairBob",
		11: "LongHairBun",
		12: "LongHairCurly",
		13: "LongHairCurvy",
		14: "LongHairDreads",
		15: "LongHairFrida",
		16: "LongHairFro",
		17: "LongHairFroBand",
		18: "LongHairNotTooLong",
		19: "LongHairShavedSides",
		20: "LongHairMiaWallace",
		21: "LongHairStraight",
		22: "LongHairStraight2",
		23: "LongHairStraightStrand",
		24: "ShortHairDreads01",
		25: "ShortHairDreads02",
		26: "ShortHairFrizzle",
		27: "ShortHairShaggyMullet",
		28: "ShortHairShortCurly",
		29: "ShortHairShortFlat",
		30: "ShortHairShortRound",
		31: "ShortHairShortWaved",
		32: "ShortHairSides",
		33: "ShortHairTheCaesar",
		34: "ShortHairTheCaesarSidePart",
	}

	s, _ := hex.DecodeString(hash)
	g.Top = m[int(s[5])%34]
}

func (g *GetAvataaarsCom) GetAvatar(user string) string {

	h := md5.New()
	io.WriteString(h, user)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	g.GenAvatarStyle(hash)
	g.GenTop(hash)
	g.GenAccessories(hash)
	g.GenHairColor(hash)
	g.GenFacialHair(hash)
	g.GenFacialHairColor(hash)
	g.GenClothes(hash)
	g.GenColorFabric(hash)
	g.GenEyes(hash)
	g.GenEyebrow(hash)
	g.GenMouth(hash)
	g.GenSkin(hash)

	return g.GenUrl()
}

func (g *GetAvataaarsCom) GenUrl() (ans string) {

	ans += fmt.Sprint("https://avataaars.io/?")
	ans += "avatarStyle=" + g.AvatarStyle
	ans += "&topType=" + g.Top
	ans += "&accessoriesType=" + g.Accessories
	ans += "&facialHairType=" + g.FacialHair
	ans += "&facialHairColor=" + g.FacialHairColor
	ans += "&clotheType=" + g.Clothes
	ans += "&clotheColor=" + g.ColorFabric
	ans += "&eyeType=" + g.Eyes
	ans += "&eyebrowType=" + g.Eyebrow
	ans += "&mouthType=" + g.Mouth
	ans += "&skinColor=" + g.Skin

	return ans
}
