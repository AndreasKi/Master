window.onload = function() {
    document.getElementById('prev_text').innerText = document.getElementById('textBox').innerText;
    setInterval(httpPostAsync, 500);
};
function httpPostAsync(){
    try {
        if (document.getElementById('textBox').innerText != document.getElementById('prev_text').innerText) {
            //var d = new Date();
            //var ms = d.getMilliseconds();
            //var s = d.getSeconds();
            var xmlHttp = new XMLHttpRequest();
            xmlHttp.open("POST", "editor_<!--ObjID--!>", true);
            var to_send = findDifference();
            to_send.diff = to_send.diff.replace(/;/g, "<--endl-->");
            var params = "text_content="+to_send.diff+"&diff_point="+ to_send.diff_start.toString() + "&equal_point="+ to_send.diff_end.toString();
            xmlHttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
            xmlHttp.setRequestHeader("Content-length", params.length);
            xmlHttp.setRequestHeader("Connection", "close");
            xmlHttp.send(params);
            //document.getElementById("julenisse").innerHTML = s + "." +ms;
        }
    }
    catch (err) {
        alert(err.message);
    }
};

function findDifference() {
    try {
        var c_text = document.getElementById('textBox').innerText;
        var p_text = document.getElementById('prev_text').innerText;
        var max_len = Math.max(c_text.length, p_text.length);

        //Find the point at which the strings deviate from eachother
        var diff_point;
        var i;
        for (i = 0; i < max_len; i++) {
            if (i < c_text.length && i < p_text.length) {
                if (c_text.substring(0, i) != p_text.substring(0, i)) {
                    diff_point = i - 1;
                    break;
                }
            } else {
                diff_point = i - 1;
                break;
            }
        }

        //Find the point at which the strings have an equal coda
        var c_equal_point;
        var p_equal_point;
        var ci = c_text.length - 1;
        var pi = p_text.length - 1;
        while (true) {
            if (pi <= diff_point) {
                c_equal_point = ci + 1;
                p_equal_point = diff_point + 1;
                break;
            } else if (ci <= diff_point) {
                c_equal_point = diff_point + 1;
                p_equal_point = pi + 1;
                break;
            } else if (c_text.substring(ci, c_text.length-1) != p_text.substring(pi, p_text.length-1)) {
                c_equal_point = ci + 1;
                p_equal_point = pi + 1;
                break;
            }
            ci--;
            pi--;
        }

        //Determine if when c overwrites p, we delete from p, or add to p
        var difference;
        if (c_equal_point > diff_point) {
            difference = c_text.substring(diff_point, c_equal_point);
        } else {
            difference = "";
        }

        if (QMLNeedsMoreChars(difference)) { //Check if we need to add chars
            var neg = 0; //How many extra chars to overwrite to ensure QML handles the change in negative direction
            var pos = 0; //How many extra chars to overwrite to ensure QML handles the change in positive direction
            var n;
            for (n = 1; n > 0; n--) {
                neg = n;
                var c_char = c_text.substring(diff_point-n , diff_point-(n-1));
                if (c_char != " " && c_char != "\n") {
                    break;
                }
            }
            var p;
            for (p = 1; p < c_text.length-c_equal_point+1; p++) {
                pos = p;
                var c_char = c_text.substring(c_equal_point+(p-1) , c_equal_point+p);
                if (c_char != " " && c_char != "\n") {
                    break;
                }
            }
            difference = c_text.substring(diff_point-neg,diff_point) + difference + c_text.substring(c_equal_point,c_equal_point+pos);
            diff_point = diff_point - neg;
            p_equal_point = p_equal_point + pos;
        }

        document.getElementById('prev_text').innerText = c_text;

        return {
            diff: difference,
            diff_start: diff_point,
            diff_end: p_equal_point
        };
    }
    catch (err) {
        alert(err.message);
    }
};

function QMLNeedsMoreChars(difference_string) {
        var outcome = false;
        var chars = difference_string.split("");
        if (chars[0] == " " || chars[0] == "\n") {
            outcome = true;
        } else if (chars[chars.length-1] == " " || chars[chars.length-1] == "\n") {
            outcome = true;
        }
        return outcome;
}

function getSelectionCharOffsetsWithin(element) {
    var start = 0;
    var sel, range, priorRange;
    if (typeof window.getSelection != "undefined") {
        range = window.getSelection().getRangeAt(0);
        priorRange = range.cloneRange();
        priorRange.selectNodeContents(element);
        priorRange.setEnd(range.startContainer, range.startOffset);
        start = priorRange.toString().length;
    } else if (typeof document.selection != "undefined" && (sel = document.selection).type != "Control") {
        range = sel.createRange();
        priorRange = document.body.createTextRange();
        priorRange.moveToElementText(element);
        priorRange.setEndPoint("EndToStart", range);
        start = priorRange.text.length;
    }
    return start-36;
};
function getSelectedText() {
    var sel_text = "";
    if (typeof window.getSelection != "undefined") {
        sel_text = window.getSelection().toString();
    } else if (typeof document.selection != "undefined" && document.selection.type != "Control") {
        sel_text = document.selection.createRange().text;
    }
    return sel_text;
};
function makeBold() {
	document.getElementById('selectedtext_bold').value = getSelectedText();
    document.getElementById('pos_bold').value = getSelectionCharOffsetsWithin(document.getElementById('textBox'));
	document.getElementById('fulltext_bold').value = document.getElementById('textBox').innerText;
	document.getElementById('makebold').submit();
};
function makeItalic() {
    document.getElementById('selectedtext_italic').value = getSelectedText();
    document.getElementById('pos_italic').value = getSelectionCharOffsetsWithin(document.getElementById('textBox'));
    document.getElementById('fulltext_italic').value = document.getElementById('textBox').innerText;
    document.getElementById('makeitalic').submit();
};
function makeUnderlined() {
    document.getElementById('selectedtext_underlined').value = getSelectedText();
    document.getElementById('pos_underlined').value = getSelectionCharOffsetsWithin(document.getElementById('textBox'));
    document.getElementById('fulltext_underlined').value = document.getElementById('textBox').innerText;
    document.getElementById('makeunderlined').submit();
};
function exitEditor() {
    try {
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function() {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                if (xmlHttp.responseText == "ack") {
                    window.open('', '_self', ''); 
                    window.close();
                } else {
                    alert("Error: Failed to close application!");
                    window.open('', '_self', ''); 
                    window.close();
                }
            }
        };
        var object_id = document.getElementById('obj_id').innerText;
        xmlHttp.open("GET", "exit_"+object_id, true);
        xmlHttp.send();
    }
    catch (err) {
        alert(err.message);
    }
}