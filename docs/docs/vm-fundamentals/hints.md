---
sidebar_position: 3
---

# Hints

There are several hints, explain each of them here. How they interact with the VM and affects it.

## Dictionaries

A dictionary in the Cairo VM is represented by a memory segment. This segment is tracked in a `ZeroDictionaryManager` which exists in a scope managed by the `ScopeManager`

`ZeroDictionaryManager` is a mapping of a segment index to a `ZeroDictionary`. There can be different dictionary managers in different scopes.

`ZeroDictionary` has 3 fields: 
1. `Data`: map storing the (key, value) data pairs of the dictionary.
2. `DefaultValue`: optional field which holds the default value of a key if it doesn't exist in the `Data` field.
3. `FreeOffset`: tracks the next free offset in the dictionary segment.  

Dictionary operations in cario are covered in two library files:
1. [dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/dict.cairo)
2. [default_dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/default_dict.cairo)

The functions here require certain hints to be implemented in the VM.

**Dict functions:**
1. dict_new
2. default_dict_new
3. dict_read
4. dict_write
5. dict_update
6. dict_squash

**Hint usage in dict functions:**

| Hint                              | Usage                 |
|-----------------------------------|-----------------------|
| DictNew                           | dict_new, dict_squash |
| DefaultDictNew                    | default_dict_new      |
| DictRead                          | dict_read             |
| DictWrite                         | dict_write            |
| DictUpdate                        | dict_update           |
| DictSquashCopyDict                | dict_squash           |
| VMExitScope                       | dict_squash           |
| SquashDict                        | dict_squash           |
| SquashDictInnerFirstIteration     | dict_squash           |
| SquashDictInnerSkipLoop           | dict_squash           |
| SquashDictInnerCheckAccessIndex   | dict_squash           |
| SquashDictInnerContinueLoop       | dict_squash           |
| SquashDictInnerLenAssert          | dict_squash           |
| SquashDictInnerUsedAccessesAssert | dict_squash           |
| SquashDictInnerAssertLenKeys      | dict_squash           |
| SquashDictInnerNextKey            | dict_squash           |
| DictSquashUpdatePtr               | dict_squash           |

**dict_new**

Creates a new dictionary. It expects a `initial_dict` variable set in the scope. The `DictNew` hint in the function creates a new `DictionaryManager` in the scope if not present and uses the `initial_dict` scope variable to seed and create a new dictionary. 

**default_dict_new**

Creates a new dictionary, with a default value. It expects a `default_value` cairo variable. The `DefaultDictNew` hint in the function creates a new `DictionaryManager` in the scope if not present and creates a new dictionary where the default value of a key which is not present returns the `default_value` cairo variable.

**dict_read**

Reads a value from the dictionary and returns the result. Reading a key from a dictionary involves writing 3 values to the dictionary segment: key, prev_value and new_value. The `DictRead` hint in the function takes care of incrementing the `current_ptr` of the dictionary in the `DictionaryManager` by 3 and writing the read value to a cairo variable `value`.

**dict_write**

Writes a value to the dictionary, overriding the existing value. Writing a key to a dictionary involves writing 3 values to the dictionary segment: key, prev_value and new_value. The `DictWrite` hint in the function takes care of incrementing the `current_ptr` of the dictionary in the `DictionaryManager` by 3, updating the actual value of the key in the `DictionaryManager` and writing the `prev_value` value to the dictionary segment.

**dict_update**

Updates a value in a dict. Updating a key in a dictionary involves writing 3 values to the dictionary segment: key, prev_value and new_value. The `DictUpdate` hint in the function takes care of incrementing the `current_ptr` of the dictionary in the `DictionaryManager` by 3, updating the actual value of the key in the `DictionaryManager` and asserting that the given cairo variable `prev_value` matches the value of the key before updating.

**dict_squash**

Returns a new dictionary with one DictAccess instance per key (value before and value after) which summarizes all the changes to that key. 13 hints are involved as illustrated in the table above.
