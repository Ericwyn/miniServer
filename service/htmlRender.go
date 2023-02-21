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
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, shrink-to-fit=no"/>
    <meta name="renderer" content="webkit"/>
    <meta name="force-rendering" content="webkit"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1"/>
    <meta name="referrer" content="never">
	<style>

	</style>
</head>

<body>
<pre style="font-size:16px">` + conf.VersionStr + `</pre>
<br>
<table style="font-size:20px; text-align: left;">
    <thead>
    <tr class="header" id="theader">
        <th>&nbsp;&nbsp;名称</th>
        <th>&nbsp;&nbsp;大小</th>
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
<br>
<br>
<br>
<footer style="text-align: left">
    <span>
		&nbsp;&nbsp;Run By
		<a href="https://github.com/Ericwyn/miniServer" target="_blank" rel="nofollow noopener">
		MiniServer
		</a>
    </span>
</footer>
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
