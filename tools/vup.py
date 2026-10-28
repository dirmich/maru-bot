"""프로젝트의 현재 버전 표기를 함께 올린다. 과거 이력은 보존한다."""
import pathlib
import re
import sys

root = pathlib.Path(__file__).resolve().parents[1]
version_file = root / 'pkg/config/version.go'
old = re.search(r'const Version = "([\d.]+)"', version_file.read_text(encoding='utf-8')).group(1)
new = sys.argv[1] if len(sys.argv) > 1 else '.'.join(old.split('.')[:2] + [str(int(old.split('.')[2]) + 1)])
if not re.fullmatch(r'\d+\.\d+\.\d+', new):
    raise SystemExit('올바른 버전을 입력하세요')
for file in [version_file, root/'cmd/marubot/main.go', *root.glob('README*.md')]:
    content = file.read_text(encoding='utf-8')
    file.write_text(content.replace(old, new), encoding='utf-8', newline='\n')
frontend = root/'admin/frontend/version.js'
frontend.parent.mkdir(parents=True, exist_ok=True)
frontend.write_text(f'export const version = "{new}";\n', encoding='utf-8')
print(new)
