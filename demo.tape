# Output file configuration
Output demo.gif

# Terminal Appearance & Styling
Set Shell "bash"
Set Theme "Catppuccin Macchiato"
Set FontSize 15
Set Width 1200
Set Height 700
Set Padding 20
Set Margin 20
Set MarginFill "#1e1e2e"
Set WindowBar Colorful
Set TypingSpeed 60ms

# Setup temporary demo workspace
Hide
Type "rm -rf /tmp/vhs-demo && mkdir -p /tmp/vhs-demo && cd /tmp/vhs-demo" Enter
Type "export PATH=$PATH:/Users/rasulkhan/projects/flagd/bin" Enter
Type "clear" Enter
Show

# --- STEP 1: Initialize Flagctl Workspace ---
Type "# 1. Initialize flagctl workspace" Enter
Sleep 500ms
Type "flagctl init --language typescript" Enter
Sleep 2s

Type "" Enter
Type "" Enter
# --- STEP 2: Create a Feature Flag ---
Type "# 2. Create typed feature flag (auto-generates TypeScript accessors)" Enter
Sleep 500ms
Type "flagctl create --key new-checkout-flow --type boolean --default on --description 'New checkout flow'" Enter
Sleep 2s

Type "" Enter
Type "" Enter
# --- STEP 3: Add Rule Targeting (Allowlist & Denylist) ---
Type "# 3. Add Allowlist & Denylist targeting rules" Enter
Sleep 500ms
Type "flagctl target add --key new-checkout-flow --attribute email --operator endsWith --value '@company.com' --variant on" Enter
Sleep 1.5s

Type "" Enter
Type "flagctl target add --key new-checkout-flow --attribute email --operator endsWith --value '@competitor.com' --variant off --top" Enter
Sleep 2s

Type "" Enter
Type "" Enter
# --- STEP 4: Progressive Feature Launch & Ramping ---
Type "# 4. Start 20% canary launch & ramp up to 50%" Enter
Sleep 500ms
Type "flagctl launch add --key new-checkout-flow --percent 20 --variant on" Enter
Sleep 1.5s

Type "" Enter
Type "flagctl launch ramp --key new-checkout-flow --percent 50" Enter
Sleep 2s

Type "" Enter
Type "" Enter
# --- STEP 5: Inspect Ordered Rule Chain ---
Type "# 5. Inspect ordered rule chain & overlap hints" Enter
Sleep 500ms
Type "flagctl target list --key new-checkout-flow" Enter
Sleep 2.5s

Type "" Enter
Type "" Enter
# --- STEP 6: Validate & Audit Codebase ---
Type "# 6. Validate JSON schema & audit codebase health" Enter
Sleep 500ms
Type "flagctl validate && flagctl audit" Enter
Sleep 3s
