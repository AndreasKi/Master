//filename: text_ops.go
//information: created on 3rd of November 2015 by Andreas Kittilsland
package main

import (
	//"fmt"
	"gopkg.in/qml.v1"
	"strings"
)

//Make selected text bold
func (ctrl *Control) Make_bold() {
	if ctrl.Startselection > -1 && ctrl.Endselection > 0 {
		startFragment, selection, endFragment := FormatInput(ctrl)

		//Check if we want to add tags, or remove tags
		if strings.Contains(selection, "font-weight:600;") {
			//Remove bold tags
			selection = strings.Replace(selection, "font-weight:600;", "", -1)
			ctrl.Plaintext = startFragment + selection + endFragment
		} else {
			//Add bold tags
			ctrl.Plaintext = startFragment + "<b>" + selection + "</b>" + endFragment
		}
		qml.Changed(ctrl, &ctrl.Plaintext)
	}
}

//Make selected text italic
func (ctrl *Control) Make_italic() {
	if ctrl.Startselection > -1 && ctrl.Endselection > 0 {
		startFragment, selection, endFragment := FormatInput(ctrl)

		//Check if we want to add tags, or remove tags
		if strings.Contains(selection, "italic;") {
			//Remove italic tags
			selection = strings.Replace(selection, "italic;", "", -1)
			ctrl.Plaintext = startFragment + selection + endFragment
		} else {
			//Add italic tags
			ctrl.Plaintext = startFragment + "<i>" + selection + "</i>" + endFragment
		}
		qml.Changed(ctrl, &ctrl.Plaintext)
	}
}

//Underline selected text
func (ctrl *Control) Make_underlined() {
	if ctrl.Startselection > -1 && ctrl.Endselection > 0 {
		startFragment, selection, endFragment := FormatInput(ctrl)

		//Check if we want to add tags, or remove tags
		if strings.Contains(selection, "text-decoration: underline;") {
			//Remove underline tags
			selection = strings.Replace(selection, "text-decoration: underline;", "", -1)
			ctrl.Plaintext = startFragment + selection + endFragment
		} else {
			//Add underline tags
			ctrl.Plaintext = startFragment + "<u>" + selection + "</u>" + endFragment
		}
		qml.Changed(ctrl, &ctrl.Plaintext)
	}
}

//Format the input so we can work with it
func FormatInput(ctrl *Control) (string, string, string) {
	pre := ctrl.Plainbeforetext
	post := ctrl.Plainaftertext
	sel := ctrl.Plainselectedtext

	//Check if the text was formatted already
	temp := strings.Split(sel, "<html>")
	if len(temp) > 1 {
		//Get the substrings of the formatting. We want to preserve already set formatting
		if len(pre) > 0 {
			pre = strings.Split(pre, "<!--EndFragment-->")[0]
		}

		sel = strings.Split(sel, "<!--StartFragment-->")[1]
		sel = strings.Split(sel, "<!--EndFragment-->")[0]

		//Avoid seg fault if selection included the last character
		if len(post) != 0 {
			post_a := strings.Split(post, "<!--StartFragment-->")
			if len(post_a) > 1 {
				post = post_a[1]
			} else {
				post = post_a[0]
			}

		}
	}
	//fmt.Println(sel)
	return pre, sel, post
}
