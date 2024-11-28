// ImportSecp256R1P hint imports the `SECP_P` constant from SECP256R1 curve utilities in the current scope

func main() {
    %{ from starkware.cairo.common.cairo_secp.secp256r1_utils import SECP256R1_P as SECP_P %}

    return();
}
