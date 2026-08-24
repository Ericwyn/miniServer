(function () {
    "use strict";

    var root = document.documentElement;
    var themeToggle = document.getElementById("theme-toggle");
    var themeColor = document.getElementById("theme-color");
    var darkMedia = window.matchMedia ? window.matchMedia("(prefers-color-scheme: dark)") : null;
    var savedTheme = null;

    try {
        savedTheme = window.localStorage.getItem("miniserver-theme");
    } catch (_) {
        savedTheme = null;
    }

    function systemTheme() {
        return darkMedia && darkMedia.matches ? "dark" : "light";
    }

    function activeTheme() {
        return root.getAttribute("data-theme") || systemTheme();
    }

    function syncThemeChrome() {
        var isDark = activeTheme() === "dark";
        if (themeToggle) {
            themeToggle.setAttribute("aria-label", isDark ? "切换到浅色主题" : "切换到深色主题");
            themeToggle.setAttribute("title", isDark ? "切换到浅色主题" : "切换到深色主题");
        }
        if (themeColor) {
            themeColor.setAttribute("content", isDark ? "#101612" : "#f1eee5");
        }
    }

    function setTheme(theme, persist) {
        root.setAttribute("data-theme", theme);
        if (persist) {
            try {
                window.localStorage.setItem("miniserver-theme", theme);
            } catch (_) {
                // Storage can be unavailable in strict privacy modes.
            }
        }
        syncThemeChrome();
    }

    if (savedTheme === "light" || savedTheme === "dark") {
        setTheme(savedTheme, false);
    } else {
        syncThemeChrome();
    }

    if (themeToggle) {
        themeToggle.addEventListener("click", function () {
            setTheme(activeTheme() === "dark" ? "light" : "dark", true);
        });
    }

    if (darkMedia) {
        var handleSystemTheme = function () {
            if (!root.hasAttribute("data-theme")) {
                syncThemeChrome();
            }
        };
        if (typeof darkMedia.addEventListener === "function") {
            darkMedia.addEventListener("change", handleSystemTheme);
        } else if (typeof darkMedia.addListener === "function") {
            darkMedia.addListener(handleSystemTheme);
        }
    }

    var fileList = document.getElementById("file-list");
    var searchInput = document.getElementById("file-search");
    var sortSelect = document.getElementById("file-sort");
    var resultCount = document.getElementById("result-count");
    var searchEmpty = document.getElementById("search-empty");
    var rows = fileList ? Array.prototype.slice.call(fileList.querySelectorAll(".file-row")) : [];
    var collator = null;
    try {
        if (typeof Intl !== "undefined" && Intl.Collator) {
            collator = new Intl.Collator(["zh-CN", "en"], { numeric: true, sensitivity: "base" });
        }
    } catch (_) {
        collator = null;
    }

    function compareText(left, right) {
        if (collator) {
            return collator.compare(left, right);
        }
        return left.localeCompare(right);
    }

    function sortRows() {
        if (!fileList) {
            return;
        }
        var mode = sortSelect ? sortSelect.value : "name";
        rows.sort(function (left, right) {
            var leftKind = left.getAttribute("data-kind");
            var rightKind = right.getAttribute("data-kind");
            if (leftKind !== rightKind) {
                return leftKind === "dir" ? -1 : 1;
            }

            if (mode === "time") {
                var timeDifference = Number(right.getAttribute("data-time")) - Number(left.getAttribute("data-time"));
                if (timeDifference !== 0) {
                    return timeDifference;
                }
            } else if (mode === "size") {
                var sizeDifference = Number(right.getAttribute("data-size")) - Number(left.getAttribute("data-size"));
                if (sizeDifference !== 0) {
                    return sizeDifference;
                }
            } else if (mode === "type") {
                var typeDifference = compareText(left.getAttribute("data-type"), right.getAttribute("data-type"));
                if (typeDifference !== 0) {
                    return typeDifference;
                }
            }

            return compareText(left.getAttribute("data-name"), right.getAttribute("data-name"));
        });

        var fragment = document.createDocumentFragment();
        rows.forEach(function (row) {
            fragment.appendChild(row);
        });
        fileList.appendChild(fragment);
    }

    function filterRows() {
        var query = searchInput ? searchInput.value.trim().toLocaleLowerCase() : "";
        var visible = 0;
        rows.forEach(function (row) {
            var haystack = (row.getAttribute("data-name") + " " + row.getAttribute("data-type")).toLocaleLowerCase();
            var matched = !query || haystack.indexOf(query) !== -1;
            row.hidden = !matched;
            if (matched) {
                visible += 1;
            }
        });

        if (resultCount) {
            resultCount.textContent = String(visible);
        }
        if (searchEmpty) {
            searchEmpty.hidden = visible !== 0;
        }
    }

    if (searchInput) {
        searchInput.addEventListener("input", filterRows);
        searchInput.addEventListener("keydown", function (event) {
            if (event.key === "Escape" && searchInput.value) {
                searchInput.value = "";
                filterRows();
            }
        });
    }

    if (sortSelect) {
        sortSelect.addEventListener("change", function () {
            sortRows();
            filterRows();
        });
    }

    document.addEventListener("keydown", function (event) {
        var target = event.target;
        var tagName = target && target.tagName ? target.tagName.toLowerCase() : "";
        var isTyping = tagName === "input" || tagName === "textarea" || tagName === "select" || (target && target.isContentEditable);
        if (event.key === "/" && !isTyping && searchInput && !event.metaKey && !event.ctrlKey && !event.altKey) {
            event.preventDefault();
            searchInput.focus();
        }
    });

    var copyButton = document.getElementById("copy-link");
    var copyLabel = document.getElementById("copy-label");
    var copyResetTimer = 0;

    function fallbackCopy(text) {
        var textarea = document.createElement("textarea");
        textarea.value = text;
        textarea.setAttribute("readonly", "");
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.select();
        var copied = false;
        try {
            copied = document.execCommand("copy");
        } catch (_) {
            copied = false;
        }
        document.body.removeChild(textarea);
        return copied;
    }

    function showCopyResult(copied) {
        if (!copyButton || !copyLabel) {
            return;
        }
        window.clearTimeout(copyResetTimer);
        copyButton.classList.toggle("is-copied", copied);
        copyLabel.textContent = copied ? "已复制" : "复制失败";
        copyButton.setAttribute("aria-label", copied ? "当前目录地址已复制" : "复制当前目录地址失败");
        copyResetTimer = window.setTimeout(function () {
            copyButton.classList.remove("is-copied");
            copyLabel.textContent = "复制地址";
            copyButton.setAttribute("aria-label", "复制当前目录地址");
        }, 1800);
    }

    if (copyButton) {
        copyButton.addEventListener("click", function () {
            var currentURL = window.location.href;
            if (navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
                navigator.clipboard.writeText(currentURL).then(function () {
                    showCopyResult(true);
                }, function () {
                    showCopyResult(fallbackCopy(currentURL));
                });
            } else {
                showCopyResult(fallbackCopy(currentURL));
            }
        });
    }

    var previewDialog = document.getElementById("preview-dialog");
    var previewTitle = document.getElementById("preview-title");
    var previewStatus = document.getElementById("preview-status");
    var previewNotice = document.getElementById("preview-notice");
    var previewNoticeText = previewNotice ? previewNotice.querySelector("span") : null;
    var previewBody = document.getElementById("preview-body");
    var previewContent = document.getElementById("preview-content");
    var previewCopyButton = document.getElementById("preview-copy");
    var previewCopyLabel = document.getElementById("preview-copy-label");
    var previewCloseButton = document.getElementById("preview-close");
    var previewOpenLink = document.getElementById("preview-open");
    var previewDownloadLink = document.getElementById("preview-download");
    var previewText = "";
    var previewRequestID = 0;
    var previewController = null;
    var previewCopyResetTimer = 0;
    var previewLineLimit = 2000;
    var previewByteLimit = 4 * 1024 * 1024;

    function showPreviewDialog() {
        if (!previewDialog || previewDialog.open) {
            return;
        }
        if (typeof previewDialog.showModal === "function") {
            previewDialog.showModal();
        } else {
            previewDialog.setAttribute("open", "");
        }
        document.body.classList.add("preview-open");
    }

    function stopPreviewRequest() {
        previewRequestID += 1;
        if (previewController) {
            previewController.abort();
            previewController = null;
        }
    }

    function closePreviewDialog() {
        stopPreviewRequest();
        if (previewDialog && previewDialog.open) {
            if (typeof previewDialog.close === "function") {
                previewDialog.close();
            } else {
                previewDialog.removeAttribute("open");
            }
        }
        document.body.classList.remove("preview-open");
    }

    function resetPreviewCopyButton() {
        window.clearTimeout(previewCopyResetTimer);
        if (!previewCopyButton || !previewCopyLabel) {
            return;
        }
        previewCopyButton.classList.remove("is-copied");
        previewCopyButton.setAttribute("aria-label", "复制预览全文");
        previewCopyLabel.textContent = "复制全文";
    }

    function setPreviewError(message) {
        previewText = "";
        if (previewBody) {
            previewBody.classList.add("is-error");
            previewBody.setAttribute("aria-busy", "false");
        }
        if (previewContent) {
            previewContent.textContent = message;
        }
        if (previewStatus) {
            previewStatus.textContent = "预览失败";
        }
        if (previewNotice) {
            previewNotice.hidden = true;
        }
        if (previewCopyButton) {
            previewCopyButton.disabled = true;
        }
    }

    function parseContentRange(value) {
        var match = /^bytes\s+(\d+)-(\d+)\/(\d+|\*)$/i.exec(value || "");
        if (!match) {
            return null;
        }
        return {
            end: Number(match[2]),
            total: match[3] === "*" ? null : Number(match[3])
        };
    }

    function renderPreviewText(text, response, typeLabel, displaySize) {
        var normalized = text.replace(/\r\n?/g, "\n");
        var lines = normalized === "" ? [] : normalized.split("\n");
        if (normalized.slice(-1) === "\n") {
            lines.pop();
        }

        var truncatedByLines = lines.length > previewLineLimit;
        var contentRange = parseContentRange(response.headers.get("Content-Range"));
        var truncatedByBytes = !!(contentRange && contentRange.total !== null && contentRange.end+1 < contentRange.total);
        var visibleLines = lines.slice(0, previewLineLimit);
        previewText = visibleLines.join("\n");

        if (previewBody) {
            previewBody.classList.remove("is-error");
            previewBody.setAttribute("aria-busy", "false");
            previewBody.scrollTop = 0;
            previewBody.scrollLeft = 0;
        }
        if (previewContent) {
            previewContent.textContent = previewText;
        }
        if (previewStatus) {
            var lineText = visibleLines.length === 0 ? "空文件" : visibleLines.length + " 行";
            previewStatus.textContent = typeLabel + " · " + displaySize + " · " + lineText + (truncatedByLines || truncatedByBytes ? " · 已截断" : "");
        }

        if (previewNotice && previewNoticeText) {
            previewNotice.hidden = !(truncatedByLines || truncatedByBytes);
            if (truncatedByLines) {
                previewNoticeText.textContent = "文件超过 2000 行，仅展示前 2000 行。";
            } else if (truncatedByBytes) {
                previewNoticeText.textContent = "文件内容超过 4 MiB，当前预览已截断。";
            }
        }
        if (previewCopyButton) {
            previewCopyButton.disabled = false;
        }
    }

    function openTextPreview(row) {
        if (!previewDialog || !row) {
            return;
        }

        stopPreviewRequest();
        var requestID = previewRequestID;
        var fileURL = row.getAttribute("href");
        var fileName = row.getAttribute("data-name") || "文本文件";
        var typeLabel = row.getAttribute("data-type") || "TEXT";
        var displaySize = row.getAttribute("data-display-size") || "未知大小";

        previewText = "";
        resetPreviewCopyButton();
        if (previewTitle) {
            previewTitle.textContent = fileName;
        }
        if (previewStatus) {
            previewStatus.textContent = typeLabel + " · " + displaySize + " · 正在读取";
        }
        if (previewNotice) {
            previewNotice.hidden = true;
        }
        if (previewBody) {
            previewBody.classList.remove("is-error");
            previewBody.setAttribute("aria-busy", "true");
            previewBody.scrollTop = 0;
            previewBody.scrollLeft = 0;
        }
        if (previewContent) {
            previewContent.textContent = "正在读取文件…";
        }
        if (previewCopyButton) {
            previewCopyButton.disabled = true;
        }
        if (previewOpenLink) {
            previewOpenLink.setAttribute("href", fileURL);
        }
        if (previewDownloadLink) {
            previewDownloadLink.setAttribute("href", fileURL);
            previewDownloadLink.setAttribute("download", fileName);
        }

        showPreviewDialog();
        if (typeof AbortController === "function") {
            previewController = new AbortController();
        }
        var requestOptions = {
            cache: "no-store",
            credentials: "same-origin",
            headers: { Range: "bytes=0-" + (previewByteLimit - 1) }
        };
        if (previewController) {
            requestOptions.signal = previewController.signal;
        }

        fetch(fileURL, requestOptions).then(function (response) {
            if (requestID !== previewRequestID) {
                return null;
            }
            if (!response.ok && response.status !== 206) {
                throw new Error("HTTP " + response.status);
            }
            var contentLength = Number(response.headers.get("Content-Length"));
            if (response.status === 200 && contentLength > previewByteLimit) {
                throw new Error("服务器未返回分段内容");
            }
            return response.text().then(function (text) {
                return { response: response, text: text };
            });
        }).then(function (result) {
            if (!result || requestID !== previewRequestID) {
                return;
            }
            previewController = null;
            renderPreviewText(result.text, result.response, typeLabel, displaySize);
        }).catch(function (error) {
            if (requestID !== previewRequestID || (error && error.name === "AbortError")) {
                return;
            }
            previewController = null;
            setPreviewError("无法读取这个文件的文本预览。\n你仍然可以使用底部按钮打开或下载原文件。");
        });
    }

    if (fileList && previewDialog) {
        fileList.addEventListener("click", function (event) {
            var target = event.target;
            var row = target && typeof target.closest === "function" ? target.closest('.file-row[data-previewable="true"]') : null;
            if (!row || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
                return;
            }
            event.preventDefault();
            openTextPreview(row);
        });
    }

    if (previewCloseButton) {
        previewCloseButton.addEventListener("click", closePreviewDialog);
    }

    if (previewDialog) {
        previewDialog.addEventListener("cancel", function (event) {
            event.preventDefault();
            closePreviewDialog();
        });
        previewDialog.addEventListener("click", function (event) {
            if (event.target === previewDialog) {
                closePreviewDialog();
            }
        });
        previewDialog.addEventListener("close", function () {
            stopPreviewRequest();
            document.body.classList.remove("preview-open");
        });
    }

    if (previewCopyButton) {
        previewCopyButton.addEventListener("click", function () {
            var showResult = function (copied) {
                window.clearTimeout(previewCopyResetTimer);
                previewCopyButton.classList.toggle("is-copied", copied);
                previewCopyButton.setAttribute("aria-label", copied ? "预览内容已复制" : "复制预览内容失败");
                previewCopyLabel.textContent = copied ? "已复制" : "复制失败";
                previewCopyResetTimer = window.setTimeout(resetPreviewCopyButton, 1600);
            };
            if (navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
                navigator.clipboard.writeText(previewText).then(function () {
                    showResult(true);
                }, function () {
                    showResult(fallbackCopy(previewText));
                });
            } else {
                showResult(fallbackCopy(previewText));
            }
        });
    }

    filterRows();
}());
