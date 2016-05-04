//filename: graphics.qml
//information: created on 2nd of November 2015 by Andreas Kittilsland
import QtQuick 2.2
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.0
import QtQuick.Dialogs 1.1

ApplicationWindow {
    id: root
    visible: true
    title: "SimpleTxt - " + ctrl.filename
    property int margin: 11
    width: 1024
    height: 768

    onClosing: {    
        ctrl.exit()
    }

    toolBar: ToolBar {
        height: 50
        RowLayout {
            anchors.fill: parent
            ToolButton {
                id: newButton
                iconSource: "res/new.png"
                onClicked: {
                    if (textBox.getText (0, textBox.length) != "") {
                        newMessageDialog.open()
                    } else {
                        ctrl.new_file()
                        openFileDialog.path = ""
                    }
                }
            }
            ToolButton {
                id: openButton
                iconSource: "res/open.png"
                onClicked: {
                    if (textBox.getText (0, textBox.length) != "") {
                        openMessageDialog.open()
                    } else {
                        openFileDialog.open()
                    }
                }
            }
            ToolButton {
                id: saveButton
                iconSource: "res/save.png"
                onClicked: {
                    if (openFileDialog.path != "") {
                        saveFileDialog.path = openFileDialog.path
                        ctrl.lock()
                        ctrl.save_file()
                    } 
                    else {
                        ctrl.lock()
                        saveFileDialog.open()
                    }
                }
            }
            ToolButton {
                id: boldButton
                iconSource: "res/bold.png"
                onClicked: {
                    textBox.plainbeforetext = textBox.getFormattedText (0, textBox.selectionStart) 
                    textBox.plainselectedtext = textBox.getFormattedText (textBox.selectionStart, textBox.selectionEnd)
                    textBox.plainaftertext = textBox.getFormattedText (textBox.selectionEnd, textBox.length)  
                    ctrl.make_bold()
                }
            }
            ToolButton {
                id: italicButton
                iconSource: "res/italic.png"
                onClicked: {
                    textBox.plainbeforetext = textBox.getFormattedText (0, textBox.selectionStart) 
                    textBox.plainselectedtext = textBox.getFormattedText (textBox.selectionStart, textBox.selectionEnd)
                    textBox.plainaftertext = textBox.getFormattedText (textBox.selectionEnd, textBox.length)  
                    ctrl.make_italic()
                }
            }
            ToolButton {
                id: underlineButton
                iconSource: "res/underline.png"
                onClicked: {
                    textBox.plainbeforetext = textBox.getFormattedText (0, textBox.selectionStart) 
                    textBox.plainselectedtext = textBox.getFormattedText (textBox.selectionStart, textBox.selectionEnd)
                    textBox.plainaftertext = textBox.getFormattedText (textBox.selectionEnd, textBox.length)  
                    ctrl.make_underlined()
                }
            }
            Button {
                id: plainRichButton
                text: "Plain Text"
                onClicked: {
                    if (text == "Plain Text") {
                        text = "Rich Text" 
                        textBox.textFormat = Text.PlainText
                    } else {
                        text = "Plain Text"
                        textBox.textFormat = Text.RichText
                    }
                }
            }           
            Item { Layout.fillWidth: true }
            Button {
                text: "Minimize"
                Layout.alignment: Qt.AlignRight

                onClicked: {
                    if (text == "Minimize") {
                        toolBar.height = 30
                        text = "▾" 
                        saveButton.visible = false
                        openButton.visible = false
                        newButton.visible = false
                        boldButton.visible = false
                        italicButton.visible = false
                        underlineButton.visible = false
                        plainRichButton.visible = false
                    } else {
                        toolBar.height = 50
                        text = "Minimize"
                        saveButton.visible = true
                        openButton.visible = true
                        newButton.visible = true
                        boldButton.visible = true
                        italicButton.visible = true
                        underlineButton.visible = true
                        plainRichButton.visible = true
                    }
                }
            }
        }
    }

    statusBar: StatusBar {
        RowLayout {
            anchors.fill: parent
            Label { 
                id: logText
                text: ctrl.log 
            }
            Binding { target: ctrl; property: "log"; value: logText.text }
        }
    }

    ColumnLayout {
        id: mainLayout
        x: margin
        height: parent.height - margin
        width: parent.width - 2 * margin
        anchors.margins: margin

        //Show notifications and status in log
        GroupBox {
            id: textGroupBox
            title: "Text:"
            Layout.fillWidth: true
            Layout.fillHeight: true

            TextArea {
                id: textBox
                objectName: "textBox"
                text: ctrl.plaintext
                textFormat: Text.AutoText
                activeFocusOnPress: true
                selectByMouse: true
                //cursorPosition: ctrl.cursorposition
                readOnly: false
                height: parent.height
                width: parent.width
                property string plainbeforetext: ""
                property string plainselectedtext: ""
                property string plainaftertext: ""
            }
            Binding { target: ctrl; property: "plaintext"; value: textBox.text }
            Binding { target: ctrl; property: "plainbeforetext"; value: textBox.plainbeforetext }
            Binding { target: ctrl; property: "plainselectedtext"; value: textBox.plainselectedtext }
            Binding { target: ctrl; property: "plainaftertext"; value: textBox.plainaftertext }
            Binding { target: ctrl; property: "startselection"; value: textBox.selectionStart }
            Binding { target: ctrl; property: "endselection"; value: textBox.selectionEnd }
            Binding { target: ctrl; property: "cursorposition"; value: textBox.cursorPosition }
        }
        
    }

    FileDialog {
        id: openFileDialog
        title: "Please choose a file"
        nameFilters: [ "SimpleTxt files (*.stxt)"]
        property string path: ""
        onAccepted: {
            path = fileUrl
            ctrl.open_file()
            close()
        }
        onRejected: {
            close()
        }
    }
    Binding { target: ctrl; property: "openpath"; value: openFileDialog.path }

    FileDialog {
        id: saveFileDialog
        title: "Browse to folder to save in"
        selectExisting: false
        property string path: ""
        onAccepted: {
            path = fileUrl
            ctrl.save_file()
            close()
        }
        onRejected: {
            close()
        }
    }
    Binding { target: ctrl; property: "savepath"; value: saveFileDialog.path }

    MessageDialog {
        id: newMessageDialog
        title: "Save changes?"
        text: "Do you wish to save changes?"
        standardButtons: StandardButton.Yes | StandardButton.No
        onYes: {
            saveButton.clicked()
            openFileDialog.path = ""
            ctrl.new_file()
            close()
        }
        onNo: {
            openFileDialog.path = ""
            ctrl.new_file()
            close()
        }
    }

    MessageDialog {
        id: openMessageDialog
        title: "Save changes?"
        text: "Do you wish to save changes?"
        standardButtons: StandardButton.Yes | StandardButton.No
        onYes: {
            saveButton.clicked()
            openFileDialog.open()
            close()
        }
        onNo: {
            openFileDialog.open()
            close()
        }
    }
}
