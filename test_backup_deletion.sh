#!/bin/bash

echo "Testing BYC Backup Deletion Functionality"
echo "=========================================="
echo

# Check if backups exist
echo "Current backups in ./backups/:"
ls -la backups/
echo

# Test the backup deletion functionality
echo "Running BYC CLI and testing backup deletion..."
echo "Steps:"
echo "1. Select option 7 (Backup & Restore)"
echo "2. Select option 4 (Delete Backup)"
echo "3. Should see list of available backups"
echo "4. Enter backup name to delete"
echo

# Create a simple test input
cat > test_input.txt << EOF
7
4
test-backup
5
11
EOF

echo "Test input created. You can run:"
echo "cat test_input.txt | ./bin/byc"
echo
echo "Or run interactively:"
echo "./bin/byc"
echo
echo "Then navigate to:"
echo "7 -> Backup & Restore"
echo "4 -> Delete Backup"
echo "You should see the list of available backups before being prompted to enter the backup name." 