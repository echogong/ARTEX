"use client";

import * as React from "react";

import { InfoIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";

export function UpdateCard() {
  const [version, setVersion] = React.useState("");

  React.useEffect(() => {
    let alive = true;
    api
      .health()
      .then((h) => {
        if (alive) setVersion((h.version ?? "").replace(/^v(?=\d)/, ""));
      })
      .catch(() => {
        if (alive) setVersion("");
      });
    return () => {
      alive = false;
    };
  }, []);

  return (
    <Card className="mb-4 break-inside-avoid md:mb-6">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <InfoIcon className="size-4" />
          版本
        </CardTitle>
        <CardDescription>在线更新已关闭。程序不会检查、下载或安装新版本。</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex flex-wrap items-center gap-2 text-sm">
          <span className="text-muted-foreground">当前版本</span>
          <Badge variant="secondary" className="font-mono">
            {version || "…"}
          </Badge>
        </div>
        <p className="text-muted-foreground text-xs">
          升级请在本机手动完成：<span className="font-mono">git pull</span> 后重新编译，或执行{" "}
          <span className="font-mono">docker compose build</span> 再{" "}
          <span className="font-mono">docker compose up -d</span>
          。Postgres 数据卷、<span className="font-mono">./data</span> 与 <span className="font-mono">./skills</span>{" "}
          会保留。
        </p>
      </CardContent>
    </Card>
  );
}
