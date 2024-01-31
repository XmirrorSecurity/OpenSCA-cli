import os
import shutil
import gitlab
from gitlab.v4.objects import Project, ProjectBranch

# repo: https://github.com/python-gitlab/python-gitlab
# doc: https://python-gitlab.readthedocs.io/en/stable/api-usage.html
# pkg: pip install --upgrade python-gitlab


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
        ref = branch.get_id()
        if ref is None:
            return
        # TODO download repo to dir

    def scan(self, path: str, out: str, log: str = "opensca.log"):
        os.system(f"{self.cli} -path {path} -out {out} -log {log}")


if __name__ == "__main__":
    cli = "your-opensca-path"
    # config gitlab auth
    gl = gitlab.Gitlab(url="your_gitlab_url")
    download_dir = "./download/"
    report_dir = "./report/"
    report_ext = [".html", ".json"]
    s = OpenscaGitlabScanner(cli, gl)
    # foreach repo
    for repo, branches in s.projects().items():
        # foreach repo branch
        for branch in branches:
            pid = str(repo.get_id())
            bid = str(branch.get_id())
            project_dir = os.path.join(download_dir, pid, bid)
            output_dir = os.path.join(report_dir, pid)
            output_files = [os.path.join(output_dir, bid + ext) for ext in report_ext]
            output_log = os.path.join(output_dir, bid + ".opensca.log")
            os.makedirs(project_dir)
            os.makedirs(output_dir)
            # download repo branch
            s.download(repo, branch, project_dir)
            # scan repo
            s.scan(project_dir, ",".join(output_files), output_log)
            # delete download data
            shutil.rmtree(project_dir)
