find ./builtin_tests -type f -name "*.small*" -exec sh -c '
for file do
    newfile="${file%.small*}__small${file#*.small}"
    mv "$file" "$newfile"
done
' sh {} +
