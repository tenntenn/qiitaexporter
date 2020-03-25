package main

import "text/template"

var tmpl = template.Must(template.New("template").Parse(`+++ 
date = "{{.Date}}"
title = "{{.Title}}"
slug = "qiita-{{.ID}}" 
tags = [{{.AllTags}}]
categories = []
+++

*この記事は[Qiita]({{.URL}})の記事をエクスポートしたものです。内容が古くなっている可能性があります。*

{{.Body}}`))
