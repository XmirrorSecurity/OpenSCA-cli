import os
import shutil

# repo: https://github.com/python-gitlab/python-gitlab
# doc: https://python-gitlab.readthedocs.io/en/stable/api-usage.html
# pkg: pip install --upgrade python-gitlab
import gitlab
from gitlab.v4.objects import Project, ProjectBranch


class OpenscaGitlabScanner:

    def __init__(self, cli: str, gl: gitlab.Gitlab):
        self.cli = cli
        self.gl = gl

    def projects(self) -> dict[Project, list[ProjectBranch]]:
        projects = {}
        for p in self.gl.projects.list():
            pid = p.get_id()
            if pid is None:
                continue
            project = self.gl.projects.get(pid)
            bids = [b.get_id() for b in project.branches.list()]
            branches = [project.branches.get(bid) for bid in bids if bid != None]
            projects[project] = branches
        return projects

    def download(self, project: Project, branch: ProjectBranch, dir: str):
        ref = str(branch.get_id())
        files = project.repository_tree(path="/", ref=ref, all=True)
        for file in files:
            if file["type"] != "blob":
                continue
            filepath = file["path"]
            if not self.support_file(filepath):
                continue
            wfilepath = os.path.join(dir, filepath)
            mkdirs(os.path.dirname(wfilepath))
            with open(wfilepath, "wb") as wf:
                project.files.raw(
                    file_path=filepath,
                    ref=ref,
                    streamed=True,
                    action=wf.write,
                )
                wf.flush()

    @staticmethod
    def support_file(filename: str) -> bool:
        if filename.endswith("pom.xml"):
            return True
        return False

    def scan(self, path: str, out: str, log: str = "opensca.log"):
        os.system(f"{self.cli} -path {path} -out {out} -log {log}")


def mkdirs(dir: str):
    if not os.path.exists(dir):
        os.makedirs(dir)


def scan_gitlab(
    cli: str = "opensca-cli",
    gitlab_url: str = "http://localhost:9000",
    gitlab_token: str = "",
    download_dir: str = "download",
    report_dir: str = "report",
    report_ext: list[str] = [".html", ".json"],
    **gitlab_args,
):
    """
    cli: opensca cmd or executable file path
    gitlab_url: gitlab url
    gitlab_token: gitlab private token
    download_dir: temp dir to download repository
    report_dir: opensca report dir
    report_ext: opensca report format
    gitlab_args: gitlab client args
    """

    # config gitlab auth
    gl = gitlab.Gitlab(
        url=gitlab_url,
        private_token=gitlab_token,
        keep_base_url=True,
        **gitlab_args,
    )

    # foreach repo
    s = OpenscaGitlabScanner(cli, gl)
    for repo, branches in s.projects().items():
        # foreach repo branch
        for branch in branches:
            pid = str(repo.get_id())
            bid = str(branch.get_id())
            repo_dir = os.path.join(download_dir, pid)
            branch_dir = os.path.join(repo_dir, bid)
            output_dir = os.path.join(report_dir, pid)
            output_files = [os.path.join(output_dir, bid + ext) for ext in report_ext]
            output_log = os.path.join(output_dir, bid + ".opensca.log")
            mkdirs(branch_dir)
            mkdirs(output_dir)
            # download repo branch
            print(f"download repo:{pid} branch:{bid}")
            s.download(repo, branch, branch_dir)
            # scan repo branch
            print(f"scan repo:{pid} branch:{bid}")
            s.scan(branch_dir, ",".join(output_files), output_log)
            # delete download repo
            shutil.rmtree(repo_dir)


if __name__ == "__main__":
    scan_gitlab(
        gitlab_url="gitlab_url",
        gitlab_token="gitlab_token",
    )
