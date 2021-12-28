// Auto-generated code. Do not edit!

package main

type PacketIDMap = map[int32]string
type DirectionMap = map[Direction]PacketIDMap
type StateMap = map[GameState]DirectionMap

var PacketNames StateMap = StateMap{
	Handshaking: {
		ServerBound: {
			0x00: "SetProtocol",
		},
	},
	Login: {
		ClientBound: {
			0x00: "Disconnect",
			0x01: "EncryptionBegin",
			0x02: "Success",
			0x03: "SetCompression",
		},
		ServerBound: {
			0x00: "Start",
			0x01: "EncryptionBegin",
		},
	},
	Play: {
		ClientBound: {
			0x00: "SpawnEntity",
			0x01: "SpawnEntityExperienceOrb",
			0x10: "MultiBlockChange",
			0x11: "Transaction",
			0x12: "CloseWindow",
			0x13: "OpenWindow",
			0x14: "WindowItems",
			0x15: "WindowData",
			0x16: "SetSlot",
			0x17: "SetCooldown",
			0x18: "CustomPayload",
			0x19: "CustomSoundEffect",
			0x1a: "KickDisconnect",
			0x1b: "EntityStatus",
			0x1c: "Explosion",
			0x1d: "UnloadChunk",
			0x1e: "GameStateChange",
			0x1f: "KeepAlive",
			0x02: "SpawnEntityWeather",
			0x20: "MapChunk",
			0x21: "WorldEvent",
			0x22: "WorldParticles",
			0x23: "Login",
			0x24: "Map",
			0x25: "Entity",
			0x26: "RelEntityMove",
			0x27: "RelEntityMoveLook",
			0x28: "EntityLook",
			0x29: "VehicleMove",
			0x2a: "OpenSignEditor",
			0x2b: "AutoRecipe",
			0x2c: "Abilities",
			0x2d: "CombatEvent",
			0x2e: "PlayerInfo",
			0x2f: "Position",
			0x03: "SpawnEntityLiving",
			0x30: "Bed",
			0x31: "Recipes",
			0x32: "EntityDestroy",
			0x33: "RemoveEntityEffect",
			0x34: "ResourcePackSend",
			0x35: "Respawn",
			0x36: "EntityHeadRotation",
			0x37: "SelectAdvancementTab",
			0x38: "WorldBorder",
			0x39: "Camera",
			0x3a: "HeldItemSlot",
			0x3b: "ScoreboardDisplayObjective",
			0x3c: "EntityMetadata",
			0x3d: "AttachEntity",
			0x3e: "EntityVelocity",
			0x3f: "EntityEquipment",
			0x04: "SpawnEntityPainting",
			0x40: "Experience",
			0x41: "UpdateHealth",
			0x42: "ScoreboardObjective",
			0x43: "Mount",
			0x44: "ScoreboardTeam",
			0x45: "ScoreboardScore",
			0x46: "SpawnPosition",
			0x47: "UpdateTime",
			0x48: "Title",
			0x49: "NamedSoundEffect",
			0x4a: "PlayerListHeaderFooter",
			0x4b: "Collect",
			0x4c: "EntityTeleport",
			0x4d: "Advancements",
			0x4e: "UpdateAttributes",
			0x4f: "EntityEffect",
			0x05: "NamedEntitySpawn",
			0x06: "Animation",
			0x07: "Statistic",
			0x08: "BlockBreakAnimation",
			0x09: "TileEntityData",
			0x0a: "BlockAction",
			0x0b: "BlockChange",
			0x0c: "Boss",
			0x0d: "ServerDifficulty",
			0x0e: "TabComplete",
			0x0f: "Chat",
		},
		ServerBound: {
			0x00: "TeleportAccept",
			0x01: "TabComplete",
			0x10: "VehicleMove",
			0x11: "BoatMove",
			0x12: "AutoRecipe",
			0x13: "Abilities",
			0x14: "BlockDig",
			0x15: "EntityAction",
			0x16: "SteerVehicle",
			0x17: "RecipeDisplayed",
			0x18: "ResourcePackStatus",
			0x19: "Advancements",
			0x1a: "HeldItemSlot",
			0x1b: "SetCreativeSlot",
			0x1c: "UpdateSign",
			0x1d: "ArmAnimation",
			0x1e: "Spectate",
			0x1f: "UseItem",
			0x02: "Chat",
			0x20: "BlockPlace",
			0x03: "ClientCommand",
			0x04: "Settings",
			0x05: "Transaction",
			0x06: "EnchantItem",
			0x07: "WindowClick",
			0x08: "CloseWindow",
			0x09: "CustomPayload",
			0x0a: "UseEntity",
			0x0b: "KeepAlive",
			0x0c: "Flying",
			0x0d: "Position",
			0x0e: "PositionLook",
			0x0f: "Look",
		},
	},
	Status: {
		ClientBound: {
			0x00: "ServerInfo",
			0x01: "Pong",
		},
		ServerBound: {
			0x00: "Start",
			0x01: "Ping",
		},
	},
}

const OHandshakingSetProtocol = 0x00
const ILoginDisconnect = 0x00
const ILoginEncryptionBegin = 0x01
const ILoginSuccess = 0x02
const ILoginSetCompression = 0x03
const OLoginStart = 0x00
const OLoginEncryptionBegin = 0x01
const IPlaySpawnEntity = 0x00
const IPlaySpawnEntityExperienceOrb = 0x01
const IPlayMultiBlockChange = 0x10
const IPlayTransaction = 0x11
const IPlayCloseWindow = 0x12
const IPlayOpenWindow = 0x13
const IPlayWindowItems = 0x14
const IPlayWindowData = 0x15
const IPlaySetSlot = 0x16
const IPlaySetCooldown = 0x17
const IPlayCustomPayload = 0x18
const IPlayCustomSoundEffect = 0x19
const IPlayKickDisconnect = 0x1a
const IPlayEntityStatus = 0x1b
const IPlayExplosion = 0x1c
const IPlayUnloadChunk = 0x1d
const IPlayGameStateChange = 0x1e
const IPlayKeepAlive = 0x1f
const IPlaySpawnEntityWeather = 0x02
const IPlayMapChunk = 0x20
const IPlayWorldEvent = 0x21
const IPlayWorldParticles = 0x22
const IPlayLogin = 0x23
const IPlayMap = 0x24
const IPlayEntity = 0x25
const IPlayRelEntityMove = 0x26
const IPlayRelEntityMoveLook = 0x27
const IPlayEntityLook = 0x28
const IPlayVehicleMove = 0x29
const IPlayOpenSignEditor = 0x2a
const IPlayAutoRecipe = 0x2b
const IPlayAbilities = 0x2c
const IPlayCombatEvent = 0x2d
const IPlayPlayerInfo = 0x2e
const IPlayPosition = 0x2f
const IPlaySpawnEntityLiving = 0x03
const IPlayBed = 0x30
const IPlayRecipes = 0x31
const IPlayEntityDestroy = 0x32
const IPlayRemoveEntityEffect = 0x33
const IPlayResourcePackSend = 0x34
const IPlayRespawn = 0x35
const IPlayEntityHeadRotation = 0x36
const IPlaySelectAdvancementTab = 0x37
const IPlayWorldBorder = 0x38
const IPlayCamera = 0x39
const IPlayHeldItemSlot = 0x3a
const IPlayScoreboardDisplayObjective = 0x3b
const IPlayEntityMetadata = 0x3c
const IPlayAttachEntity = 0x3d
const IPlayEntityVelocity = 0x3e
const IPlayEntityEquipment = 0x3f
const IPlaySpawnEntityPainting = 0x04
const IPlayExperience = 0x40
const IPlayUpdateHealth = 0x41
const IPlayScoreboardObjective = 0x42
const IPlayMount = 0x43
const IPlayScoreboardTeam = 0x44
const IPlayScoreboardScore = 0x45
const IPlaySpawnPosition = 0x46
const IPlayUpdateTime = 0x47
const IPlayTitle = 0x48
const IPlayNamedSoundEffect = 0x49
const IPlayPlayerListHeaderFooter = 0x4a
const IPlayCollect = 0x4b
const IPlayEntityTeleport = 0x4c
const IPlayAdvancements = 0x4d
const IPlayUpdateAttributes = 0x4e
const IPlayEntityEffect = 0x4f
const IPlayNamedEntitySpawn = 0x05
const IPlayAnimation = 0x06
const IPlayStatistic = 0x07
const IPlayBlockBreakAnimation = 0x08
const IPlayTileEntityData = 0x09
const IPlayBlockAction = 0x0a
const IPlayBlockChange = 0x0b
const IPlayBoss = 0x0c
const IPlayServerDifficulty = 0x0d
const IPlayTabComplete = 0x0e
const IPlayChat = 0x0f
const OPlayTeleportAccept = 0x00
const OPlayTabComplete = 0x01
const OPlayVehicleMove = 0x10
const OPlayBoatMove = 0x11
const OPlayAutoRecipe = 0x12
const OPlayAbilities = 0x13
const OPlayBlockDig = 0x14
const OPlayEntityAction = 0x15
const OPlaySteerVehicle = 0x16
const OPlayRecipeDisplayed = 0x17
const OPlayResourcePackStatus = 0x18
const OPlayAdvancements = 0x19
const OPlayHeldItemSlot = 0x1a
const OPlaySetCreativeSlot = 0x1b
const OPlayUpdateSign = 0x1c
const OPlayArmAnimation = 0x1d
const OPlaySpectate = 0x1e
const OPlayUseItem = 0x1f
const OPlayChat = 0x02
const OPlayBlockPlace = 0x20
const OPlayClientCommand = 0x03
const OPlaySettings = 0x04
const OPlayTransaction = 0x05
const OPlayEnchantItem = 0x06
const OPlayWindowClick = 0x07
const OPlayCloseWindow = 0x08
const OPlayCustomPayload = 0x09
const OPlayUseEntity = 0x0a
const OPlayKeepAlive = 0x0b
const OPlayFlying = 0x0c
const OPlayPosition = 0x0d
const OPlayPositionLook = 0x0e
const OPlayLook = 0x0f
const IStatusServerInfo = 0x00
const IStatusPong = 0x01
const OStatusStart = 0x00
const OStatusPing = 0x01
