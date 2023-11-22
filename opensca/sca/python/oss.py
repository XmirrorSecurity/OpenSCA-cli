import re
import sys
import json

def parse_setup_py(setup_py_path):
	"""解析setup.py文件"""

	with open(setup_py_path, "r", encoding='utf8') as f:
		pass_func = lambda **x: x
		try:
			import distutils
			distutils.core.setup = pass_func # type: ignore
		except Exception:
			pass
		try:
			import setuptools
			setuptools.setup = pass_func # type: ignore
		except Exception:
			pass

		# 获取setup参数
		args = {}
		code = re.sub(r'(?<!\w)setup\(','args=setup(',f.read())
		code = code.replace('__file__','"{}"'.format(setup_py_path))
		exec(code, args)
		if 'args' in args:
			data = args['args']
			info = {}
			for k in ['name','version','license','packages','install_requires','requires']:
				if k in data:
					info[k] = data[k]
			print('opensca_start<<{}>>opensca_end'.format(json.dumps(info)))

if __name__ == "__main__":
	if len(sys.argv) > 1:
		parse_setup_py(sys.argv[1])