#!/bin/bash
set -e

# Problem:
# cdk8s-cli v2.204.x generates invalid Go code for certain K8s 1.30+ resources (e.g., APIService).
# It emits a `With(mixins ...constructs.IMixin)` method interacting with `constructs.IMixin`.
# However, `constructs.IMixin` was REMOVED in `constructs` v10, causing build failures.
#
# Solution:
# This script blindly strips the offending method declaration and implementation from the generated files.
# This is safe because we do not use the `With()` mixin functionality in our CDK8s code.

TARGET_DIR="imports"

echo "🔧 Patching generated CDK8s imports in $TARGET_DIR..."

if [ ! -d "$TARGET_DIR" ]; then
    echo "⚠️  Directory $TARGET_DIR does not exist. Skipping patch."
    exit 0
fi

# 1. Remove the interface method declaration: `With(mixins ...constructs.IMixin) constructs.IConstruct`
# We use perl for robust multiline matching.
find "$TARGET_DIR" -name "*.go" -print0 | xargs -0 perl -i -0777 -pe 's/\tWith\(mixins \.\.\.constructs\.IMixin\) constructs\.IConstruct\n//g'

# 2. Remove the implementation function: `func (k *jsiiProxy_...) With(...) ... { ... }`
# Matches from `func ... With` down to the closing brace of the function.
find "$TARGET_DIR" -name "*.go" -print0 | xargs -0 perl -i -0777 -pe 's/^func [^\n]*?With\(mixins \.\.\.constructs\.IMixin\) constructs\.IConstruct \{.*?\n^\}\n//gms'

echo "✅ Patch applied successfully. Legacy IMixin references removed."
