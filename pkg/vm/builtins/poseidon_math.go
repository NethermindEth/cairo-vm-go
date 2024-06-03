package builtins

import (
	"sync"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const (
	Base16 = 16
	Base10 = 10
)

const (
	fullRounds    = 8
	partialRounds = 83
)

func addRoundKeys(state []fp.Element, round int) {
	state[0].Add(&state[0], &roundKeys[round][0])
	state[1].Add(&state[1], &roundKeys[round][1])
	state[2].Add(&state[2], &roundKeys[round][2])
}

func subWords(state []fp.Element, full bool) {
	squared := fp.Element{}
	state[2].Mul(&state[2], squared.Mul(&state[2], &state[2]))
	if full {
		state[0].Mul(&state[0], squared.Mul(&state[0], &state[0]))
		state[1].Mul(&state[1], squared.Mul(&state[1], &state[1]))
	}
}

// M = ((3,1,1), (1,-1,1), (1,1,-2))
// state = M * state
func mixLayer(state []fp.Element) {
	stateSum := new(fp.Element).Add(&state[0], &state[1])
	stateSum.Add(stateSum, &state[2])
	state[0].Double(&state[0]).Add(&state[0], stateSum)                                // stateSum + state[0] * 2
	state[1].Sub(stateSum, state[1].Double(&state[1]))                                 // stateSum - state[1] * 2
	state[2].Sub(stateSum, state[2].Add(&state[2], new(fp.Element).Double(&state[2]))) // stateSum - state[2] * 3
}

func round(state []fp.Element, full bool, index int) {
	addRoundKeys(state, index)
	subWords(state, full)
	mixLayer(state)
}

func hadesPermutation(state []fp.Element) {
	initialiseRoundKeys.Do(setRoundKeys)
	totalRounds := fullRounds + partialRounds
	for i := 0; i < totalRounds; i++ {
		full := (i < fullRounds/2) || (totalRounds-i <= fullRounds/2)
		round(state, full, i)
	}
}

func PoseidonPerm(x, y, z *fp.Element) []fp.Element {
	state := []fp.Element{*x, *y, *z}
	hadesPermutation(state)
	return state
}

var (
	initialiseRoundKeys sync.Once
	roundKeys           = [][]fp.Element{}
)

func setRoundKeys() {
	for _, keysStr := range roundKeysSpec {
		curRound := []fp.Element{}
		for _, keyStr := range keysStr {
			key, err := new(fp.Element).SetString(keyStr)
			if err != nil {
				panic(key)
			}
			if keyStr != key.Text(Base10) {
				panic("round key not in stark field " + keyStr)
			}

			curRound = append(curRound, *key)
		}
		roundKeys = append(roundKeys, curRound)
	}
}

// https://github.com/starkware-industries/poseidon/blob/main/poseidon3.txt
var roundKeysSpec = [][]string{
	{
		"2950795762459345168613727575620414179244544320470208355568817838579231751791",
		"1587446564224215276866294500450702039420286416111469274423465069420553242820",
		"1645965921169490687904413452218868659025437693527479459426157555728339600137",
	},
	{
		"2782373324549879794752287702905278018819686065818504085638398966973694145741",
		"3409172630025222641379726933524480516420204828329395644967085131392375707302",
		"2379053116496905638239090788901387719228422033660130943198035907032739387135",
	},
	{
		"2570819397480941104144008784293466051718826502582588529995520356691856497111",
		"3546220846133880637977653625763703334841539452343273304410918449202580719746",
		"2720682389492889709700489490056111332164748138023159726590726667539759963454",
	},
	{
		"1899653471897224903834726250400246354200311275092866725547887381599836519005",
		"2369443697923857319844855392163763375394720104106200469525915896159690979559",
		"2354174693689535854311272135513626412848402744119855553970180659094265527996",
	},
	{
		"2404084503073127963385083467393598147276436640877011103379112521338973185443",
		"950320777137731763811524327595514151340412860090489448295239456547370725376",
		"2121140748740143694053732746913428481442990369183417228688865837805149503386",
	},
	{
		"2372065044800422557577242066480215868569521938346032514014152523102053709709",
		"2618497439310693947058545060953893433487994458443568169824149550389484489896",
		"3518297267402065742048564133910509847197496119850246255805075095266319996916",
	},
	{
		"340529752683340505065238931581518232901634742162506851191464448040657139775",
		"1954876811294863748406056845662382214841467408616109501720437541211031966538",
		"813813157354633930267029888722341725864333883175521358739311868164460385261",
	},
	{
		"71901595776070443337150458310956362034911936706490730914901986556638720031",
		"2789761472166115462625363403490399263810962093264318361008954888847594113421",
		"2628791615374802560074754031104384456692791616314774034906110098358135152410",
	},
	{
		"3617032588734559635167557152518265808024917503198278888820567553943986939719",
		"2624012360209966117322788103333497793082705816015202046036057821340914061980",
		"149101987103211771991327927827692640556911620408176100290586418839323044234",
	},
	{
		"1039927963829140138166373450440320262590862908847727961488297105916489431045",
		"2213946951050724449162431068646025833746639391992751674082854766704900195669",
		"2792724903541814965769131737117981991997031078369482697195201969174353468597",
	},
	{
		"3212031629728871219804596347439383805499808476303618848198208101593976279441",
		"3343514080098703935339621028041191631325798327656683100151836206557453199613",
		"614054702436541219556958850933730254992710988573177298270089989048553060199",
	},
	{
		"148148081026449726283933484730968827750202042869875329032965774667206931170",
		"1158283532103191908366672518396366136968613180867652172211392033571980848414",
		"1032400527342371389481069504520755916075559110755235773196747439146396688513",
	},
	{
		"806900704622005851310078578853499250941978435851598088619290797134710613736",
		"462498083559902778091095573017508352472262817904991134671058825705968404510",
		"1003580119810278869589347418043095667699674425582646347949349245557449452503",
	},
	{
		"619074932220101074089137133998298830285661916867732916607601635248249357793",
		"2635090520059500019661864086615522409798872905401305311748231832709078452746",
		"978252636251682252755279071140187792306115352460774007308726210405257135181",
	},
	{
		"1766912167973123409669091967764158892111310474906691336473559256218048677083",
		"1663265127259512472182980890707014969235283233442916350121860684522654120381",
		"3532407621206959585000336211742670185380751515636605428496206887841428074250",
	},
	{
		"2507023127157093845256722098502856938353143387711652912931112668310034975446",
		"3321152907858462102434883844787153373036767230808678981306827073335525034593",
		"3039253036806065280643845548147711477270022154459620569428286684179698125661",
	},
	{
		"103480338868480851881924519768416587261556021758163719199282794248762465380",
		"2394049781357087698434751577708655768465803975478348134669006211289636928495",
		"2660531560345476340796109810821127229446538730404600368347902087220064379579",
	},
	{
		"3603166934034556203649050570865466556260359798872408576857928196141785055563",
		"1553799760191949768532188139643704561532896296986025007089826672890485412324",
		"2744284717053657689091306578463476341218866418732695211367062598446038965164",
	},
	{
		"320745764922149897598257794663594419839885234101078803811049904310835548856",
		"979382242100682161589753881721708883681034024104145498709287731138044566302",
		"1860426855810549882740147175136418997351054138609396651615467358416651354991",
	},
	{
		"336173081054369235994909356892506146234495707857220254489443629387613956145",
		"1632470326779699229772327605759783482411227247311431865655466227711078175883",
		"921958250077481394074960433988881176409497663777043304881055317463712938502",
	},
	{
		"3034358982193370602048539901033542101022185309652879937418114324899281842797",
		"25626282149517463867572353922222474817434101087272320606729439087234878607",
		"3002662261401575565838149305485737102400501329139562227180277188790091853682",
	},
	{
		"2939684373453383817196521641512509179310654199629514917426341354023324109367",
		"1076484609897998179434851570277297233169621096172424141759873688902355505136",
		"2575095284833160494841112025725243274091830284746697961080467506739203605049",
	},
	{
		"3565075264617591783581665711620369529657840830498005563542124551465195621851",
		"2197016502533303822395077038351174326125210255869204501838837289716363437993",
		"331415322883530754594261416546036195982886300052707474899691116664327869405",
	},
	{
		"1935011233711290003793244296594669823169522055520303479680359990463281661839",
		"3495901467168087413996941216661589517270845976538454329511167073314577412322",
		"954195417117133246453562983448451025087661597543338750600301835944144520375",
	},
	{
		"1271840477709992894995746871435810599280944810893784031132923384456797925777",
		"2565310762274337662754531859505158700827688964841878141121196528015826671847",
		"3365022288251637014588279139038152521653896670895105540140002607272936852513",
	},
	{
		"1660592021628965529963974299647026602622092163312666588591285654477111176051",
		"970104372286014048279296575474974982288801187216974504035759997141059513421",
		"2617024574317953753849168721871770134225690844968986289121504184985993971227",
	},
	{
		"999899815343607746071464113462778273556695659506865124478430189024755832262",
		"2228536129413411161615629030408828764980855956560026807518714080003644769896",
		"2701953891198001564547196795777701119629537795442025393867364730330476403227",
	},
	{
		"837078355588159388741598313782044128527494922918203556465116291436461597853",
		"2121749601840466143704862369657561429793951309962582099604848281796392359214",
		"771812260179247428733132708063116523892339056677915387749121983038690154755",
	},
	{
		"3317336423132806446086732225036532603224267214833263122557471741829060578219",
		"481570067997721834712647566896657604857788523050900222145547508314620762046",
		"242195042559343964206291740270858862066153636168162642380846129622127460192",
	},
	{
		"2855462178889999218204481481614105202770810647859867354506557827319138379686",
		"3525521107148375040131784770413887305850308357895464453970651672160034885202",
		"1320839531502392535964065058804908871811967681250362364246430459003920305799",
	},
	{
		"2514191518588387125173345107242226637171897291221681115249521904869763202419",
		"2798335750958827619666318316247381695117827718387653874070218127140615157902",
		"2808467767967035643407948058486565877867906577474361783201337540214875566395",
	},
	{
		"3551834385992706206273955480294669176699286104229279436819137165202231595747",
		"1219439673853113792340300173186247996249367102884530407862469123523013083971",
		"761519904537984520554247997444508040636526566551719396202550009393012691157",
	},
	{
		"3355402549169351700500518865338783382387571349497391475317206324155237401353",
		"199541098009731541347317515995192175813554789571447733944970283654592727138",
		"192100490643078165121235261796864975568292640203635147901612231594408079071",
	},
	{
		"1187019357602953326192019968809486933768550466167033084944727938441427050581",
		"189525349641911362389041124808934468936759383310282010671081989585219065700",
		"2831653363992091308880573627558515686245403755586311978724025292003353336665",
	},
	{
		"2052859812632218952608271535089179639890275494426396974475479657192657094698",
		"1670756178709659908159049531058853320846231785448204274277900022176591811072",
		"3538757242013734574731807289786598937548399719866320954894004830207085723125",
	},
	{
		"710549042741321081781917034337800036872214466705318638023070812391485261299",
		"2345013122330545298606028187653996682275206910242635100920038943391319595180",
		"3528369671971445493932880023233332035122954362711876290904323783426765912206",
	},
	{
		"1167120829038120978297497195837406760848728897181138760506162680655977700764",
		"3073243357129146594530765548901087443775563058893907738967898816092270628884",
		"378514724418106317738164464176041649567501099164061863402473942795977719726",
	},
	{
		"333391138410406330127594722511180398159664250722328578952158227406762627796",
		"1727570175639917398410201375510924114487348765559913502662122372848626931905",
		"968312190621809249603425066974405725769739606059422769908547372904403793174",
	},
	{
		"360659316299446405855194688051178331671817370423873014757323462844775818348",
		"1386580151907705298970465943238806620109618995410132218037375811184684929291",
		"3604888328937389309031638299660239238400230206645344173700074923133890528967",
	},
	{
		"2496185632263372962152518155651824899299616724241852816983268163379540137546",
		"486538168871046887467737983064272608432052269868418721234810979756540672990",
		"1558415498960552213241704009433360128041672577274390114589014204605400783336",
	},
	{
		"3512058327686147326577190314835092911156317204978509183234511559551181053926",
		"2235429387083113882635494090887463486491842634403047716936833563914243946191",
		"1290896777143878193192832813769470418518651727840187056683408155503813799882",
	},
	{
		"1143310336918357319571079551779316654556781203013096026972411429993634080835",
		"3235435208525081966062419599803346573407862428113723170955762956243193422118",
		"1293239921425673430660897025143433077974838969258268884994339615096356996604",
	},
	{
		"236252269127612784685426260840574970698541177557674806964960352572864382971",
		"1733907592497266237374827232200506798207318263912423249709509725341212026275",
		"302004309771755665128395814807589350526779835595021835389022325987048089868",
	},
	{
		"3018926838139221755384801385583867283206879023218491758435446265703006270945",
		"39701437664873825906031098349904330565195980985885489447836580931425171297",
		"908381723021746969965674308809436059628307487140174335882627549095646509778",
	},
	{
		"219062858908229855064136253265968615354041842047384625689776811853821594358",
		"1283129863776453589317845316917890202859466483456216900835390291449830275503",
		"418512623547417594896140369190919231877873410935689672661226540908900544012",
	},
	{
		"1792181590047131972851015200157890246436013346535432437041535789841136268632",
		"370546432987510607338044736824316856592558876687225326692366316978098770516",
		"3323437805230586112013581113386626899534419826098235300155664022709435756946",
	},
	{
		"910076621742039763058481476739499965761942516177975130656340375573185415877",
		"1762188042455633427137702520675816545396284185254002959309669405982213803405",
		"2186362253913140345102191078329764107619534641234549431429008219905315900520",
	},
	{
		"2230647725927681765419218738218528849146504088716182944327179019215826045083",
		"1069243907556644434301190076451112491469636357133398376850435321160857761825",
		"2695241469149243992683268025359863087303400907336026926662328156934068747593",
	},
	{
		"1361519681544413849831669554199151294308350560528931040264950307931824877035",
		"1339116632207878730171031743761550901312154740800549632983325427035029084904",
		"790593524918851401449292693473498591068920069246127392274811084156907468875",
	},
	{
		"2723400368331924254840192318398326090089058735091724263333980290765736363637",
		"3457180265095920471443772463283225391927927225993685928066766687141729456030",
		"1483675376954327086153452545475557749815683871577400883707749788555424847954",
	},
	{
		"2926303836265506736227240325795090239680154099205721426928300056982414025239",
		"543969119775473768170832347411484329362572550684421616624136244239799475526",
		"237401230683847084256617415614300816373730178313253487575312839074042461932",
	},
	{
		"844568412840391587862072008674263874021460074878949862892685736454654414423",
		"151922054871708336050647150237534498235916969120198637893731715254687336644",
		"1299332034710622815055321547569101119597030148120309411086203580212105652312",
	},
	{
		"487046922649899823989594814663418784068895385009696501386459462815688122993",
		"1104883249092599185744249485896585912845784382683240114120846423960548576851",
		"1458388705536282069567179348797334876446380557083422364875248475157495514484",
	},
	{
		"850248109622750774031817200193861444623975329881731864752464222442574976566",
		"2885843173858536690032695698009109793537724845140477446409245651176355435722",
		"3027068551635372249579348422266406787688980506275086097330568993357835463816",
	},
	{
		"3231892723647447539926175383213338123506134054432701323145045438168976970994",
		"1719080830641935421242626784132692936776388194122314954558418655725251172826",
		"1172253756541066126131022537343350498482225068791630219494878195815226839450",
	},
	{
		"1619232269633026603732619978083169293258272967781186544174521481891163985093",
		"3495680684841853175973173610562400042003100419811771341346135531754869014567",
		"1576161515913099892951745452471618612307857113799539794680346855318958552758",
	},
	{
		"2618326122974253423403350731396350223238201817594761152626832144510903048529",
		"2696245132758436974032479782852265185094623165224532063951287925001108567649",
		"930116505665110070247395429730201844026054810856263733273443066419816003444",
	},
	{
		"2786389174502246248523918824488629229455088716707062764363111940462137404076",
		"1555260846425735320214671887347115247546042526197895180675436886484523605116",
		"2306241912153325247392671742757902161446877415586158295423293240351799505917",
	},
	{
		"411529621724849932999694270803131456243889635467661223241617477462914950626",
		"1542495485262286701469125140275904136434075186064076910329015697714211835205",
		"1853045663799041100600825096887578544265580718909350942241802897995488264551",
	},
	{
		"2963055259497271220202739837493041799968576111953080503132045092194513937286",
		"2303806870349915764285872605046527036748108533406243381676768310692344456050",
		"2622104986201990620910286730213140904984256464479840856728424375142929278875",
	},
	{
		"2369987021925266811581727383184031736927816625797282287927222602539037105864",
		"285070227712021899602056480426671736057274017903028992288878116056674401781",
		"3034087076179360957800568733595959058628497428787907887933697691951454610691",
	},
	{
		"469095854351700119980323115747590868855368701825706298740201488006320881056",
		"360001976264385426746283365024817520563236378289230404095383746911725100012",
		"3438709327109021347267562000879503009590697221730578667498351600602230296178",
	},
	{
		"63573904800572228121671659287593650438456772568903228287754075619928214969",
		"3470881855042989871434874691030920672110111605547839662680968354703074556970",
		"724559311507950497340993415408274803001166693839947519425501269424891465492",
	},
	{
		"880409284677518997550768549487344416321062350742831373397603704465823658986",
		"6876255662475867703077362872097208259197756317287339941435193538565586230",
		"2701916445133770775447884812906226786217969545216086200932273680400909154638",
	},
	{
		"425152119158711585559310064242720816611629181537672850898056934507216982586",
		"1475552998258917706756737045704649573088377604240716286977690565239187213744",
		"2413772448122400684309006716414417978370152271397082147158000439863002593561",
	},
	{
		"392160855822256520519339260245328807036619920858503984710539815951012864164",
		"1075036996503791536261050742318169965707018400307026402939804424927087093987",
		"2176439430328703902070742432016450246365760303014562857296722712989275658921",
	},
	{
		"1413865976587623331051814207977382826721471106513581745229680113383908569693",
		"4879283427490523253696177116563427032332223531862961281430108575019551814",
		"3392583297537374046875199552977614390492290683707960975137418536812266544902",
	},
	{
		"3600854486849487646325182927019642276644093512133907046667282144129939150983",
		"2779924664161372134024229593301361846129279572186444474616319283535189797834",
		"2722699960903170449291146429799738181514821447014433304730310678334403972040",
	},
	{
		"819109815049226540285781191874507704729062681836086010078910930707209464699",
		"3046121243742768013822760785918001632929744274211027071381357122228091333823",
		"1339019590803056172509793134119156250729668216522001157582155155947567682278",
	},
	{
		"1933279639657506214789316403763326578443023901555983256955812717638093967201",
		"2138221547112520744699126051903811860205771600821672121643894708182292213541",
		"2694713515543641924097704224170357995809887124438248292930846280951601597065",
	},
	{
		"2471734202930133750093618989223585244499567111661178960753938272334153710615",
		"504903761112092757611047718215309856203214372330635774577409639907729993533",
		"1943979703748281357156510253941035712048221353507135074336243405478613241290",
	},
	{
		"684525210957572142559049112233609445802004614280157992196913315652663518936",
		"1705585400798782397786453706717059483604368413512485532079242223503960814508",
		"192429517716023021556170942988476050278432319516032402725586427701913624665",
	},
	{
		"1586493702243128040549584165333371192888583026298039652930372758731750166765",
		"686072673323546915014972146032384917012218151266600268450347114036285993377",
		"3464340397998075738891129996710075228740496767934137465519455338004332839215",
	},
	{
		"2805249176617071054530589390406083958753103601524808155663551392362371834663",
		"667746464250968521164727418691487653339733392025160477655836902744186489526",
		"1131527712905109997177270289411406385352032457456054589588342450404257139778",
	},
	{
		"1908969485750011212309284349900149072003218505891252313183123635318886241171",
		"1025257076985551890132050019084873267454083056307650830147063480409707787695",
		"2153175291918371429502545470578981828372846236838301412119329786849737957977",
	},
	{
		"3410257749736714576487217882785226905621212230027780855361670645857085424384",
		"3442969106887588154491488961893254739289120695377621434680934888062399029952",
		"3029953900235731770255937704976720759948880815387104275525268727341390470237",
	},
	{
		"85453456084781138713939104192561924536933417707871501802199311333127894466",
		"2730629666577257820220329078741301754580009106438115341296453318350676425129",
		"178242450661072967256438102630920745430303027840919213764087927763335940415",
	},
	{
		"2844589222514708695700541363167856718216388819406388706818431442998498677557",
		"3547876269219141094308889387292091231377253967587961309624916269569559952944",
		"2525005406762984211707203144785482908331876505006839217175334833739957826850",
	},
	{
		"3096397013555211396701910432830904669391580557191845136003938801598654871345",
		"574424067119200181933992948252007230348512600107123873197603373898923821490",
		"1714030696055067278349157346067719307863507310709155690164546226450579547098",
	},
	{
		"2339895272202694698739231405357972261413383527237194045718815176814132612501",
		"3562501318971895161271663840954705079797767042115717360959659475564651685069",
		"69069358687197963617161747606993436483967992689488259107924379545671193749",
	},
	{
		"2614502738369008850475068874731531583863538486212691941619835266611116051561",
		"655247349763023251625727726218660142895322325659927266813592114640858573566",
		"2305235672527595714255517865498269719545193172975330668070873705108690670678",
	},
	{
		"926416070297755413261159098243058134401665060349723804040714357642180531931",
		"866523735635840246543516964237513287099659681479228450791071595433217821460",
		"2284334068466681424919271582037156124891004191915573957556691163266198707693",
	},
	{
		"1812588309302477291425732810913354633465435706480768615104211305579383928792",
		"2836899808619013605432050476764608707770404125005720004551836441247917488507",
		"2989087789022865112405242078196235025698647423649950459911546051695688370523",
	},
	{
		"68056284404189102136488263779598243992465747932368669388126367131855404486",
		"505425339250887519581119854377342241317528319745596963584548343662758204398",
		"2118963546856545068961709089296976921067035227488975882615462246481055679215",
	},
	{
		"2253872596319969096156004495313034590996995209785432485705134570745135149681",
		"1625090409149943603241183848936692198923183279116014478406452426158572703264",
		"179139838844452470348634657368199622305888473747024389514258107503778442495",
	},
	{
		"1567067018147735642071130442904093290030432522257811793540290101391210410341",
		"2737301854006865242314806979738760349397411136469975337509958305470398783585",
		"3002738216460904473515791428798860225499078134627026021350799206894618186256",
	},
	{
		"374029488099466837453096950537275565120689146401077127482884887409712315162",
		"973403256517481077805460710540468856199855789930951602150773500862180885363",
		"2691967457038172130555117632010860984519926022632800605713473799739632878867",
	},
	{
		"3515906794910381201365530594248181418811879320679684239326734893975752012109",
		"148057579455448384062325089530558091463206199724854022070244924642222283388",
		"1541588700238272710315890873051237741033408846596322948443180470429851502842",
	},
	{
		"147013865879011936545137344076637170977925826031496203944786839068852795297",
		"2630278389304735265620281704608245039972003761509102213752997636382302839857",
		"1359048670759642844930007747955701205155822111403150159614453244477853867621",
	},
	{
		"2438984569205812336319229336885480537793786558293523767186829418969842616677",
		"2137792255841525507649318539501906353254503076308308692873313199435029594138",
		"2262318076430740712267739371170174514379142884859595360065535117601097652755",
	},
	{
		"2792703718581084537295613508201818489836796608902614779596544185252826291584",
		"2294173715793292812015960640392421991604150133581218254866878921346561546149",
		"2770011224727997178743274791849308200493823127651418989170761007078565678171",
	},
}
