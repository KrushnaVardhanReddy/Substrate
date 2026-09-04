import os
import glob

files = glob.glob("/home/krushna/Project/Substrate/dashboard/**/*.svelte", recursive=True)
count = 0
for f in files:
    with open(f, 'r') as file:
        lines = file.readlines()
    
    if len(lines) > 0 and lines[0].startswith("// Copyright"):
        start = 0
        while start < len(lines) and lines[start].startswith("//"):
            start += 1
        
        new_lines = lines[start:]
        if len(new_lines) > 0 and new_lines[0].strip() == "":
            new_lines = new_lines[1:]
            
        with open(f, 'w') as file:
            file.writelines(new_lines)
        count += 1

print(f"Fixed {count} svelte files.")
