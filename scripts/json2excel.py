import json
import pandas as pd


class Dependency:
    def __init__(self):
        self.vendor = ""
        self.name = ""
        self.version = ""
        self.language = ""
        self.direct = False
        self.paths: list[str] = []
        self.licenses: list[License] = []
        self.vulnerabilities: list[Vulnerability] = []
        pass

    @staticmethod
    def col_name():
        return [
            "Vendor",
            "Name",
            "Version",
            "Language",
            "Direct",
            "Path",
            "License",
            "Vulnerability",
        ]

    def raw(self):
        lics = ",".join([lic.name for lic in self.licenses])
        vuls = ",".join([vul.id for vul in self.vulnerabilities])
        paths = ";".join(self.paths)
        direct = "Direct" if self.direct else "Indirect"
        return [
            self.vendor,
            self.name,
            self.version,
            self.language,
            direct,
            paths,
            lics,
            vuls,
        ]


class License:
    def __init__(self):
        self.name = ""
        pass


class Vulnerability:
    def __init__(self) -> None:
        self.name = ""
        self.id = ""
        self.cve_id = ""
        self.cwe_id = ""
        self.cnvd_id = ""
        self.cnnvd_id = ""
        self.release_date = ""
        self.description = ""
        self.suggestion = ""
        # 风险等级 1:严重 2:高 3:中 4:低
        self.security_level_id = 0
        # 利用难度 0:不可利用 1:可利用
        self.exploit_level_id = 0
        pass

    @staticmethod
    def col_name():
        return [
            "Name",
            "ID",
            "CVE ID",
            "CWE ID",
            "CNVD ID",
            "CNNVD ID",
            "Release Date",
            "Description",
            "Suggestion",
            "Security Level",
            "Available",
        ]

    def raw(self):
        security_level_map = {
            0: "U",  # Unknown
            1: "C",
            2: "H",
            3: "M",
            4: "L",
        }
        security_level = security_level_map[self.security_level_id]
        available = "N" if self.exploit_level_id == 0 else "Y"
        return [
            self.name,
            self.id,
            self.cve_id,
            self.cwe_id,
            self.cnvd_id,
            self.cnnvd_id,
            self.release_date,
            self.description,
            self.suggestion,
            security_level,
            available,
        ]
        pass


def read_dep(data: dict) -> Dependency:
    dep = Dependency()
    if "vendor" in data:
        dep.vendor = data["vendor"]
    if "name" in data:
        dep.name = data["name"]
    if "version" in data:
        dep.version = data["version"]
    if "language" in data:
        dep.language = data["language"]
    if "direct" in data:
        dep.direct = data["direct"]
    if "paths" in data:
        dep.paths = data["paths"]
    if "licenses" in data:
        for license in data["licenses"]:
            lic = License()
            if "name" in license:
                lic.name = license["name"]
            dep.licenses.append(lic)
    if "vulnerabilities" in data:
        for vuln in data["vulnerabilities"]:
            dep.vulnerabilities.append(read_vuln(vuln))
    return dep


def read_vuln(data: dict) -> Vulnerability:
    vuln = Vulnerability()
    if "name" in data:
        vuln.name = data["name"]
    if "id" in data:
        vuln.id = data["id"]
    if "cve_id" in data:
        vuln.cve_id = data["cve_id"]
    if "cwe_id" in data:
        vuln.cwe_id = data["cwe_id"]
    if "cnvd_id" in data:
        vuln.cnvd_id = data["cnvd_id"]
    if "cnnvd_id" in data:
        vuln.cnnvd_id = data["cnnvd_id"]
    if "release_date" in data:
        vuln.release_date = data["release_date"]
    if "description" in data:
        vuln.description = data["description"]
    if "suggestion" in data:
        vuln.suggestion = data["suggestion"]
    if "security_level_id" in data:
        vuln.security_level_id = data["security_level_id"]
    if "exploit_level_id" in data:
        vuln.exploit_level_id = data["exploit_level_id"]
    return vuln


def read_deps(f: str) -> list[Dependency]:
    deps = []
    json_data = json.load(open(f, encoding="utf-8"))
    q = [json_data]
    while len(q) > 0:
        n = q[0]
        q = q[1:]
        dep = read_dep(n)
        if dep.name != "":
            deps.append(dep)
        if "children" not in n:
            continue
        q.extend(n["children"])
    return deps


def json2excel(input: str, output: str):
    deps = read_deps(input)
    vuls = {}
    for dep in deps:
        for vuln in dep.vulnerabilities:
            vuls[vuln.id] = vuln
    writer = pd.ExcelWriter(output)
    # 保存组件
    dep_df = pd.DataFrame([d.raw() for d in deps], columns=Dependency.col_name())
    dep_df.to_excel(writer, sheet_name="Dependencies", index=False)
    # 保存漏洞
    vul_df = pd.DataFrame([v.raw() for v in vuls.values()], columns=Vulnerability.col_name())
    vul_df.to_excel(writer, sheet_name="Vulnerabilities", index=False)
    writer.close()


if __name__ == "__main__":
    json2excel("result.json", "result.xlsx")
