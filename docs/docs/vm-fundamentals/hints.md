---
sidebar_position: 3
---

# Hints

There are several hints, explain each of them here. How they interact with the VM and affect it.

## Dictionaries

In the Cairo VM, dictionaries are represented by memory segments managed by the **ZeroDictionaryManager** within a scope handled by the **ScopeManager**.

**ZeroDictionaryManager**

A **ZeroDictionaryManager** maps a segment index to a **ZeroDictionary**. Different dictionary managers can exist in various scopes.

**ZeroDictionary**

A **ZeroDictionary** consists of three fields:

1. **Data**: A map storing the (key, value) pairs of the dictionary.
2. **DefaultValue**: An optional field holding the default value for a key if it doesn't exist in the Data field.
3. **FreeOffset**: Tracks the next free offset in the dictionary segment.

A dictionary segment writes data in sets of three values:
1. **Key**
2. **Previous Value**
3. **New Value**

For example, if key k1 has a value v1 and is updated to value v2, the dictionary segment will write three values: k1, v1, and v2 in consecutive offsets. Any dictionary access operation, such as reading or writing, will similarly add data to the segment. For instance, a read operation on key k1 with value v1 will write k1, v1, and v1 to consecutive offsets.

Dictionary operations in Cairo are covered in two library files:
1. [dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/dict.cairo)
2. [default_dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/default_dict.cairo)

These functions require specific hints to be implemented in the VM.

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

Creates a new dictionary. It requires an **initial_dict** variable set in the scope. The **DictNew** hint creates a new **DictionaryManager** in the scope if not present and uses the **initial_dict** variable to seed and create a new dictionary.

**default_dict_new**

Creates a new dictionary with a default value. It expects a **default_value** Cairo variable. The **DefaultDictNew** hint creates a new **DictionaryManager** in the scope if not present and creates a new dictionary where the default value for a key not present in the Data field is the **default_value** variable.

**dict_read**

Reads a value from the dictionary and returns it. Reading a key involves writing three values to the dictionary segment: key, previous value, and new value. The **DictRead** hint increments the current_ptr of the dictionary in the **DictionaryManager** by three and writes the read value to a Cairo variable **value**.

**dict_write**

Writes a value to the dictionary, overriding the existing value. Writing a key involves writing three values to the dictionary segment: key, previous value, and new value. The **DictWrite** hint increments the current_ptr of the dictionary in the **DictionaryManager** by three, updates the actual value of the key in the **Dictionary**, and writes the **prev_value** to the dictionary segment.

**dict_update**

Updates a value in the dictionary. Updating a key involves writing three values to the dictionary segment: key, previous value, and new value. The **DictUpdate** hint increments the current_ptr of the dictionary in the **DictionaryManager** by three, updates the actual value of the key in the **Dictionary**, and asserts that the given Cairo variable prev_value matches the key's value before updating.

**dict_squash**

Returns a new dictionary with one DictAccess instance per key (value before and value after) summarizing all the changes to that key. The dict_squash function involves 13 hints, as shown in the table above.

Example:

Input: {(key1, 0, 2), (key1, 2, 7), (key2, 4, 1), (key1, 7, 5), (key2, 1, 2)}

Output: {(key1, 0, 5), (key2, 4, 2)} 
