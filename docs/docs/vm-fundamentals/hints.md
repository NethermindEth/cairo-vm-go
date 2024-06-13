---
sidebar_position: 3
---

# Hints

There are several hints, explain each of them here. How they interact with the VM and affects it.

## Dictionaries

A dictionary in the Cairo VM is represented by a memory segment. This segment is tracked in a DictionaryManager which exists in a scope managed by the ScopeManager.

Dictionary operations in cario are covered in two library files:
1. [dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/dict.cairo)
2. [default_dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/default_dict.cairo)

The functions here require certain hints to be implemented in the VM. We will cover each hint implemented and how it's used by the relevant library functions.


