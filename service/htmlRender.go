package service

import "github.com/Ericwyn/MiniServer/conf"

type fileMsgVO struct {
	FileName string
	FileSize string
	IsDir    bool
}

func renderMsg(fileMsgList []fileMsgVO) string {
	resStr := ""
	for _, vo := range fileMsgList {
		resStr += "    <tr>\n" +
			"        <td>\n" +
			"            &nbsp;&nbsp;<a href=\"\" class=\"next\">" + vo.FileName + "</a>&nbsp;&nbsp;&nbsp;&nbsp;\n" +
			"        </td>\n" +
			"        <td>\n" +
			"            &nbsp; " + vo.FileSize + " &nbsp;\n" +
			"        </td>\n" +
			"    </tr>"
	}

	return resStr
}

func renderHtml(fileMsgList []fileMsgVO) string {
	html := `
<!DOCTYPE html>
<html lang="zh">

<head>
    <meta charset="utf-8">
    <title id="title"> MiniServer </title>
    <style>
        h1 {
            font-size: 46px
        }
    </style>
</head>

<body>
<pre>` + conf.VersionStr + `</pre>
<table style="font-size:24px">
    <thead>
    <tr class="header" id="theader">
        <th>名称</th>
        <th>大小</th>
    </tr>
    </thead>
    <tbody id="list">
    <tr id="goback" style="display: none">
        <td>
            &nbsp;&nbsp;<a href="" id="back">返回上层目录</a>&nbsp;&nbsp;&nbsp;&nbsp;
        </td>
    </tr>
    ` + renderMsg(fileMsgList) + `
</table>
</body>
<script>
    var list = document.getElementsByClassName("next");
    for (var i = 0; i < list.length; i++) {
        if (window.location.pathname !== "/") {
            list[i].href = (window.location.pathname + "/" + list[i].text).replace("//", "/");
        } else {
            list[i].href = list[i].text;
        }
    }
    if (window.location.pathname !== "/") {
        var href = window.location.href;
        if (href.endsWith("/")) {
            href = href.substr(0, href.length - 1);
        }
        var temp = href.split("/");
        document.getElementById("back").href = href.substr(0, href.length - temp[temp.length - 1].length);
        document.getElementById("goback").style.display = "inline";
    }
</script>
</html>
`

	return html
}
