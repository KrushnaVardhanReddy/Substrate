import os
import glob

def fix_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()
    
    modified = False
    
    # Fix array of tables
    if '"diff_reports"' in content and '"preview_sessions"' not in content:
        content = content.replace('"diff_reports"', '"preview_sessions", "diff_reports"')
        modified = True
        
    # Fix DELETE statements
    if 'DELETE FROM diff_reports' in content and 'preview_sessions' not in content:
        content = content.replace('pool.Exec(ctx, "DELETE FROM diff_reports")', 'pool.Exec(ctx, "DELETE FROM preview_sessions")\n\tpool.Exec(ctx, "DELETE FROM diff_reports")')
        content = content.replace('pool.Exec(ctx, "DELETE FROM diff_reports");', 'pool.Exec(ctx, "DELETE FROM preview_sessions");\n\t_, err = pool.Exec(ctx, "DELETE FROM diff_reports");')
        modified = True

    if modified:
        with open(filepath, 'w') as f:
            f.write(content)
        print(f"Fixed {filepath}")

for f in glob.glob("scripts/e2e/*_test.go"):
    fix_file(f)
