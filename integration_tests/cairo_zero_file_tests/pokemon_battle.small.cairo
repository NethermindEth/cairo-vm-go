// This file has been borrowed from https://github.com/valerobles/blockchain_based_game/blob/922f28084117037f2abb423f04858822afb0b2cb/experimental/Playground.cairo#L4

%builtins output range_check

from starkware.cairo.common.serialize import serialize_word
from starkware.cairo.common.math import unsigned_div_rem
from starkware.cairo.common.math_cmp import is_le

struct Pokemon {
    id: felt,
    hp: felt,
    atk: felt,
    init: felt,
    def: felt,
    type1: felt,
    type2: felt,
    atk1: Attack*,
    atk2: Attack*,
    // atk3: Attack*,
    // atk4: Attack*,
}
struct Attack {
    type: felt,
    damage: felt,
}

func createBisasam() -> Pokemon* {
    return (
        new Pokemon(
            id=1,
            hp=152,
            atk=111,
            init=106,
            def=111,
            type1='grass',
            type2='',
            atk1=new Attack(type='grass', damage=30),
            atk2=new Attack(type='normal', damage=40),
        )
    );
}

func createPikachu() -> Pokemon* {
    return (
        new Pokemon(
            id=25,
            hp=142,
            atk=117,
            init=156,
            def=101,
            type1='electro',
            type2='',
            atk1=new Attack(type='electro', damage=30),
            atk2=new Attack(type='normal', damage=35),
        )
    );
}

func fight{range_check_ptr: felt, output_ptr: felt*}(pkmn1: Pokemon*, pkmn2: Pokemon*) -> Pokemon* {
    // toDo : use random attacks -> use get_random()

    alloc_locals;
    // local firstIsFaster = is_le(pkmn1.init,pkmn2.init);
    local faster_pkmn: Pokemon*;
    local slower_pkmn: Pokemon*;
    if (is_le(pkmn1.init, pkmn2.init) == 0) {
        faster_pkmn = pkmn1;
        slower_pkmn = pkmn2;
    } else {
        faster_pkmn = pkmn2;
        slower_pkmn = pkmn1;
    }

    // pkmn1 is faster
    let _dmgx = attackAndGetDamage(faster_pkmn, faster_pkmn.atk1, slower_pkmn);
    local dmg = _dmgx;

    // calculate new HP
    local pkmn2_hp = slower_pkmn.hp - dmg;

    let newPok_: Pokemon* = updateHP(slower_pkmn, pkmn2_hp);
    local newPok: Pokemon* = newPok_;

    // if new HP value less 0 -> dead
    if (is_le(newPok.hp, 0) == 1) {
        // return winner
        serialize_word(newPok.hp);
        return (faster_pkmn);
    }

    let _dmgSecondFight = attackAndGetDamage(slower_pkmn, slower_pkmn.atk1, faster_pkmn);
    local dmgSecondFight = _dmgSecondFight;

    local pkmn1_hp = faster_pkmn.hp - dmg;

    let newPok_2: Pokemon* = updateHP(faster_pkmn, pkmn1_hp);
    local newPok2: Pokemon* = newPok_2;

    if (is_le(newPok2.hp, 0) == 1) {
        // return winner
        serialize_word(newPok2.hp);
        return (slower_pkmn);
    }

    let winner: Pokemon* = fight(newPok, newPok2);

    return (winner);
    //
}

func attackAndGetDamage{range_check_ptr, output_ptr: felt*}(
    pkmn1: Pokemon*, atk: Attack*, pkmn2: Pokemon*
) -> felt {
    // Damage formula = (((2* level *1 or 2) / 5  * AttackDamage * Attack.Pok1 / Defense.Pok2) / 50 )* STAB *  random (217 bis 255 / 255)
    alloc_locals;
    local stab;
    if (atk.type == pkmn1.type1) {
        stab = 2;
    }
    if (atk.type == pkmn1.type1) {
        stab = 2;
    } else {
        stab = 1;
    }

    let level = 50000;
    let rand1 = get_random(2);
    let a = 2 * level * rand1;
    let (crit, r) = unsigned_div_rem(a, 5);
    let b = crit * atk.damage * pkmn1.atk;
    let (c, r) = unsigned_div_rem(b, pkmn2.def);
    let (d, r) = unsigned_div_rem(c, 50);
    let e = d * stab;
    let f = get_random(50);
    let g = e * (f + 205);
    let (h, r) = unsigned_div_rem(g, 255);
    let (final, r) = unsigned_div_rem(h, 1000);
    // serialize_word(final);
    return (final);
}

func fightAndGetWinner(pkmn1: Pokemon, pkmn2: Pokemon) -> Pokemon {
    return (pkmn1);
}

func get_random{range_check_ptr}(range: felt) -> felt {
    let (res, r) = unsigned_div_rem(1665829291743, range);  // toDo: replace with currentTimeMillis
    return (r + 1);
}

func updateHP(pkmn: Pokemon*, hp_: felt) -> Pokemon* {
    return (
        new Pokemon(
            id=pkmn.id,
            hp=hp_,
            atk=pkmn.atk,
            init=pkmn.init,
            def=pkmn.def,
            type1=pkmn.type1,
            type2=pkmn.type2,
            atk1=pkmn.atk1,
            atk2=pkmn.atk2,
        )
    );
}

func main{output_ptr: felt*, range_check_ptr}() {
    alloc_locals;
    local bisasam: Pokemon* = createBisasam();
    local pikachu: Pokemon* = createPikachu();

    let _dmg = attackAndGetDamage(bisasam, bisasam.atk1, pikachu);
    local dmg = _dmg;
    let _dmg2 = attackAndGetDamage(pikachu, pikachu.atk1, bisasam);
    local dmg2 = _dmg2;
    let winner = fight(bisasam, pikachu);

    // serialize_word(string);
    // serialize_word(dmg);
    // serialize_word(winner.hp);

    return ();
}
