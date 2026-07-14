#!/bin/bash
mkdir -p temp_mockups

echo "Generating T14 Mockup..."
URL_T14=$(python3 scripts/stitch_submit.py --generate --project-id 4289049482274780799 --prompt "Create a global layout for an enterprise SaaS dashboard. Dark theme with deep dark background (#0F1117). Use modern sans-serif typography (Inter). Include a sidebar and a top header. The header should feature a glassmorphism effect. Use teal accent colors (#00BFA5) for primary buttons." | grep "Download URL:" | awk '{print $4}')
echo "Downloading T14..."
curl -s -o temp_mockups/t14_mockup.html "$URL_T14"
sed -i "s|{{MOCKUP_FILE_PATH}}|temp_mockups/t14_mockup.html|g" prompts/phase-11/t14_premium_aesthetics.txt

echo "Generating T10 Mockup..."
URL_T10=$(python3 scripts/stitch_submit.py --generate --project-id 4289049482274780799 --prompt "Create a Service Node card for a dependency graph. Small rectangular box, rounded corners, dark background (#1E222C). Display service name prominently, a small circular green health dot, and a language badge (e.g. Go)." | grep "Download URL:" | awk '{print $4}')
echo "Downloading T10..."
curl -s -o temp_mockups/t10_mockup.html "$URL_T10"
sed -i "s|{{MOCKUP_FILE_PATH}}|temp_mockups/t10_mockup.html|g" prompts/phase-11/t10_svelte_flow.txt

echo "Generating T11 Mockup..."
URL_T11=$(python3 scripts/stitch_submit.py --generate --project-id 4289049482274780799 --prompt "Create a Command Palette modal overlay like Raycast. Centered on a dark glassmorphism background overlay. Search input at top, list of search results below. Highlight first result. Deep dark mode." | grep "Download URL:" | awk '{print $4}')
echo "Downloading T11..."
curl -s -o temp_mockups/t11_mockup.html "$URL_T11"
sed -i "s|{{MOCKUP_FILE_PATH}}|temp_mockups/t11_mockup.html|g" prompts/phase-11/t11_command_palette.txt

echo "Generating T15 Mockup..."
URL_T15=$(python3 scripts/stitch_submit.py --generate --project-id 4289049482274780799 --prompt "Create a 3-step onboarding wizard. Dark mode design. Step 1: Connect GitHub button. Step 2: Scanning Repositories progress bar. Step 3: Enter Dashboard primary button." | grep "Download URL:" | awk '{print $4}')
echo "Downloading T15..."
curl -s -o temp_mockups/t15_mockup.html "$URL_T15"
sed -i "s|{{MOCKUP_FILE_PATH}}|temp_mockups/t15_mockup.html|g" prompts/phase-11/t15_onboarding_wizard.txt

echo "Done."
