import os

tasks_file = "tasks.md"
completed_file = "completed_tasks.md"

with open(tasks_file, "r") as f:
    lines = f.readlines()

new_tasks_lines = []
done_lines = []

for line in lines:
    if "✅ Done" in line:
        done_lines.append(line)
    else:
        new_tasks_lines.append(line)

if done_lines:
    with open(completed_file, "a") as f:
        f.writelines(done_lines)
    
    with open(tasks_file, "w") as f:
        f.writelines(new_tasks_lines)
    print(f"Moved {len(done_lines)} tasks to completed_tasks.md")
else:
    print("No completed tasks found.")
