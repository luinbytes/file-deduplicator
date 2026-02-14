#!/bin/bash
# Demo script for file-deduplicator perceptual image deduplication
# This script simulates a demo workflow for video recording

set -e

echo "=========================================="
echo "File Deduplicator v3.0.0 Demo Script"
echo "=========================================="
echo ""

# Colors for better terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

DEMO_DIR="$HOME/file-dedup-demo"
BINARY_PATH="$HOME/file-deduplicator/dist/file-deduplicator-linux-amd64"

# Step 1: Create demo directory
echo -e "${BLUE}[STEP 1]${NC} Setting up demo directory..."
rm -rf "$DEMO_DIR" 2>/dev/null || true
mkdir -p "$DEMO_DIR"
echo -e "${GREEN}✓${NC} Created demo directory: $DEMO_DIR"
echo ""

# Step 2: Check if sample images exist
echo -e "${BLUE}[STEP 2]${NC} Checking for sample images..."
SAMPLE_COUNT=$(find "$DEMO_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.jpeg" \) 2>/dev/null | wc -l)

if [ "$SAMPLE_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}⚠${NC} No sample images found!"
    echo ""
    echo "To create demo images:"
    echo "1. Add 8-10 photos to: $DEMO_DIR"
    echo "2. Include some variations (brightness, contrast, etc.)"
    echo "3. Re-run this script"
    echo ""
    exit 0
fi

echo -e "${GREEN}✓${NC} Found $SAMPLE_COUNT sample images in demo directory"
echo ""
ls -lh "$DEMO_DIR"
echo ""

# Step 3: Show problem - total file size
echo -e "${BLUE}[STEP 3]${NC} Current state: Cluttered photo library..."
TOTAL_SIZE=$(du -sh "$DEMO_DIR" | cut -f1)
echo -e "Total size: ${YELLOW}$TOTAL_SIZE${NC}"
echo -e "Number of files: ${YELLOW}$SAMPLE_COUNT${NC}"
echo ""

# Step 4: Traditional approach with fdupes (if available)
echo -e "${BLUE}[STEP 4]${NC} Traditional approach: Using fdupes..."
echo "Note: Traditional tools only find exact duplicates..."
echo ""
if command -v fdupes &> /dev/null; then
    DUPE_COUNT=$(fdupes -r "$DEMO_DIR" | grep -v "^$" | wc -l)
    echo -e "Exact duplicates found: ${YELLOW}$DUPE_COUNT${NC}"
else
    echo -e "${YELLOW}⚠${NC} fdupes not installed (install with: sudo apt install fdupes)"
fi
echo ""

# Step 5: Run file-deduplicator with perceptual mode
echo -e "${BLUE}[STEP 5]${NC} File Deduplicator: Perceptual image deduplication..."
echo ""
echo "Command: $BINARY_PATH -dir $DEMO_DIR -perceptual -similarity 10 -dry-run"
echo ""

"$BINARY_PATH" -dir "$DEMO_DIR" -perceptual -similarity 10 -dry-run
echo ""

# Step 6: Show what would happen
echo -e "${BLUE}[STEP 6]${NC} Summary of findings..."
echo ""
echo "Perceptual mode finds:"
echo "  - Exact duplicates (100% similarity)"
echo "  - Similar images with edits (brightness, contrast, filters)"
echo "  - Near-identical screenshots and burst photos"
echo ""

# Step 7: Actual cleanup (dry-run shown, actual requires confirmation)
echo -e "${BLUE}[STEP 7]${NC} Safe cleanup workflow..."
echo ""
echo "To safely clean up, run:"
echo ""
echo -e "${GREEN}  # Step 1: Preview (what you just saw above)${NC}"
echo "  $BINARY_PATH -dir $DEMO_DIR -perceptual -similarity 10 -dry-run"
echo ""
echo -e "${GREEN}  # Step 2: Move to safe folder (not delete)${NC}"
echo "  $BINARY_PATH -dir $DEMO_DIR -perceptual -similarity 10 -move-to ${DEMO_DIR}_duplicates"
echo ""
echo -e "${GREEN}  # Step 3: Review moved files${NC}"
echo "  ls ${DEMO_DIR}_duplicates"
echo ""
echo -e "${GREEN}  # Step 4: Delete only after review${NC}"
echo "  rm -rf ${DEMO_DIR}_duplicates"
echo ""

echo "=========================================="
echo "Demo complete!"
echo "=========================================="
echo ""
echo "For video recording:"
echo "1. Add sample images to: $DEMO_DIR"
echo "2. Run this script in a large terminal"
echo "3. Record screen with OBS/Loom/QuickTime"
echo "4. Follow the script prompts"
echo ""
