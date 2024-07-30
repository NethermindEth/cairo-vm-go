---
sidebar_position: 3
---

# Cairo Zero Hints

Cairo Zero relies on hints to optimize various operations. To ensure compatibility across different Cairo Virtual Machine, Starkware maintains a list of [whitelisted hints](https://github.com/starkware-libs/cairo-lang/tree/0e4dab8a6065d80d1c726394f5d9d23cb451706a/src/starkware/starknet/security/whitelists) that any Cairo VM needs to implement. These hints are essential for the proper functioning of Cairo Zero programs.

In this section, we will explore the critical role of hints in Cairo Zero, focusing on two main aspects:
- High-level operations: description of the various computations that necessitate the use of Cairo Zero hints
- Detailed hint analysis: in-depth look at all hints, explaining their purpose

## Dictionaries

In the Cairo VM, dictionaries are represented by memory segments managed by the **ZeroDictionaryManager** within a scope handled by the **ScopeManager**.

### ZeroDictionaryManager

A **ZeroDictionaryManager** maps a segment index to a **ZeroDictionary**. Different dictionary managers can exist in various scopes.

### ZeroDictionary

A **ZeroDictionary** consists of three fields:

1. **Data**: A map storing the (key, value) pairs of the dictionary.
2. **DefaultValue**: An optional field holding the default value for a key if it doesn't exist in the Data field.
3. **FreeOffset**: Tracks the next free offset in the dictionary segment.

A dictionary segment writes data in sets of three values:
1. **Key**
2. **Previous Value**
3. **New Value**

For example, if key `k1` has a value `v1` and is updated to value `v2`, the dictionary segment will write three values: `k1`, `v1`, and `v2` in consecutive offsets. Any dictionary access operation, such as reading or writing, will similarly add data to the segment. For instance, a read operation on key `k1` with value `v1` will write `k1`, `v1`, and `v1` to consecutive offsets.

Dictionary operations in Cairo are covered in two library files:
1. [dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/dict.cairo)
2. [default_dict.cairo](https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/cairo/common/default_dict.cairo)

These functions require specific hints to be implemented in the VM.

### Dict functions

1. dict_new
2. default_dict_new
3. dict_read
4. dict_write
5. dict_update
6. dict_squash

### Hint usage in dict functions

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

## Usort

`usort` Cairo [function](https://github.com/starkware-libs/cairo-lang/blob/0e4dab8a6065d80d1c726394f5d9d23cb451706a/src/starkware/cairo/common/usort.cairo#L8) is used to sort an array of field elements while removing duplicates. It returns the sorted array in ascending order and its length along with the multiplicities of each unique element.

### Hint usage in `usort` functions

Overall, `usort` requires 5 hints to be executed:

**UsortEnterScope**

Enters a new scope with `__usort_max_size` set to either:
- `0` if `__usort_max_size` variable is not found in `globals()` scope
- `1 << 20` if `__usort_max_size` variable is found `globals()` scope

This hint is used to potentially set a maximum length to the array to be sorted.

**UsortBody**

Core hint that does most of the sorting operation computation. It uses a dictionnary to group input elements and their positions Then, it sorts the unique elements and generates the output array and multiplicities.

After execution this hint, the Cairo code will call `verify_usort` recursive function, which ensures correctness of the sorting and multiplicity counting.

**UsortVerify**

Prepares for the verification of the multiplicity of the current value in the sorted output.

**UsortVerifyMultiplicityAssert**

Checks that the array of positions in scope doesn't contain any value. This hint actually implements the base case for the `verify_multiplicity` Cairo recursive function.

**UsortVerifyMultiplicityBody** 

Extracts a specific value of the sorted array with `pop`, updating indices for the verification of the next value in the recursive call.

# Cairo One Hints