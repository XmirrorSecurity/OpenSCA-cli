import sys
import json

def parse_setup_py(setup_py_path):
	"""解析setup.py文件"""
	with open(setup_py_path, "r") as f:
		pass_func = lambda **x: x
		try:
			import distutils
			distutils.core.setup = pass_func
		except Exception:
			pass
		try:
			import setuptools
			setuptools.setup = pass_func
		except Exception:
			pass
		# 获取setup参数
		args = {}
		exec(f.read().replace("setup(", "args=setup("), args)
		if 'args' in args:
			js = json.dumps(args["args"])
			print('oss_start<<{}>>oss_end'.format(js))

if __name__ == "__main__":
	if len(sys.argv) > 1:
		parse_setup_py(sys.argv[1])