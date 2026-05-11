package seed

// CategorySeed is the Go-side spec for a category. Slugs are stable
// identifiers; everything else can be edited and re-seeded.
type CategorySeed struct {
	Slug        string
	Name        string
	Description string
	Icon        string
	Color       string
}

// ActivityTypeSeed binds to a category by slug (resolved at seed time).
type ActivityTypeSeed struct {
	CategorySlug  string
	Slug          string
	Name          string
	Description   string
	XPPerUnit     int64
	PointsPerUnit int64
}

// RuleSeed binds optionally to a category by slug. Empty CategorySlug = global rule.
type RuleSeed struct {
	CategorySlug string
	Metric       string // count | xp | points | level
	Threshold    int64
}

// AchievementSeed is an entire achievement plus its rules.
type AchievementSeed struct {
	Slug        string
	Name        string
	Description string
	Title       string
	Combinator  string // ALL | ANY
	BonusPoints int64
	Rules       []RuleSeed
}

// Categories — broad household areas. Curated for an RPG-style household.
// Icons reference slugs under view/static/icons/<slug>.png.
var Categories = []CategorySeed{
	{Slug: "quehaceres", Name: "Quehaceres", Description: "Tareas del hogar y responsabilidades.", Icon: "tools", Color: "#10B981"},
	{Slug: "escuela", Name: "Escuela", Description: "Tareas escolares y aprendizaje.", Icon: "tome", Color: "#3B82F6"},
	{Slug: "higiene", Name: "Higiene", Description: "Cuidado personal y aseo.", Icon: "rose", Color: "#06B6D4"},
	{Slug: "lectura", Name: "Lectura", Description: "Hábitos de lectura y exploración de historias.", Icon: "scroll", Color: "#8B5CF6"},
	{Slug: "deporte", Name: "Deporte", Description: "Actividad física y deportes.", Icon: "bow", Color: "#F97316"},
	{Slug: "arte", Name: "Arte y Música", Description: "Creatividad, música y expresión artística.", Icon: "trumpet", Color: "#EC4899"},
	{Slug: "fe", Name: "Fe", Description: "Vida espiritual: oración, sacramentos y lectura de la Biblia.", Icon: "lamp", Color: "#D9A441"},
	{Slug: "comidas", Name: "Comidas", Description: "Comer a tiempo y con buen ánimo: desayuno, almuerzo y cena.", Icon: "cheese", Color: "#F59E0B"},
}

// ActivityTypes — what kids actually do. XP rewards effort; points reward
// optional/extra behaviors. Some types are XP-only on purpose (baseline duties).
var ActivityTypes = []ActivityTypeSeed{
	// Quehaceres
	{CategorySlug: "quehaceres", Slug: "lavar-platos", Name: "Lavar los platos", XPPerUnit: 10, PointsPerUnit: 5},
	{CategorySlug: "quehaceres", Slug: "hacer-cama", Name: "Hacer la cama", XPPerUnit: 5, PointsPerUnit: 2},
	{CategorySlug: "quehaceres", Slug: "sacar-basura", Name: "Sacar la basura", XPPerUnit: 8, PointsPerUnit: 4},
	{CategorySlug: "quehaceres", Slug: "doblar-ropa", Name: "Doblar la ropa", XPPerUnit: 10, PointsPerUnit: 5},
	{CategorySlug: "quehaceres", Slug: "aspirar", Name: "Aspirar el piso", XPPerUnit: 15, PointsPerUnit: 8},
	{CategorySlug: "quehaceres", Slug: "poner-mesa", Name: "Poner la mesa", XPPerUnit: 5, PointsPerUnit: 3},
	{CategorySlug: "quehaceres", Slug: "limpiar-cuarto", Name: "Limpiar el cuarto", XPPerUnit: 20, PointsPerUnit: 10},
	// Escuela
	{CategorySlug: "escuela", Slug: "tarea-hecha", Name: "Tarea hecha", XPPerUnit: 15, PointsPerUnit: 8},
	{CategorySlug: "escuela", Slug: "buena-nota", Name: "Sacar buena nota", XPPerUnit: 50, PointsPerUnit: 25},
	{CategorySlug: "escuela", Slug: "estudiar-examen", Name: "Estudiar para un examen", XPPerUnit: 20, PointsPerUnit: 10},
	{CategorySlug: "escuela", Slug: "presentacion", Name: "Hacer una presentación", XPPerUnit: 30, PointsPerUnit: 15},
	// Higiene — los baseline son XP-only (sin puntos) porque son obligatorios.
	{CategorySlug: "higiene", Slug: "cepillarse", Name: "Cepillarse los dientes", XPPerUnit: 3, PointsPerUnit: 0},
	{CategorySlug: "higiene", Slug: "banarse", Name: "Bañarse", XPPerUnit: 5, PointsPerUnit: 0},
	{CategorySlug: "higiene", Slug: "lavarse-manos", Name: "Lavarse las manos", XPPerUnit: 1, PointsPerUnit: 0},
	// Lectura
	{CategorySlug: "lectura", Slug: "leer-30min", Name: "Leer 30 minutos", XPPerUnit: 10, PointsPerUnit: 5},
	{CategorySlug: "lectura", Slug: "leer-1h", Name: "Leer 1 hora", XPPerUnit: 20, PointsPerUnit: 10},
	{CategorySlug: "lectura", Slug: "terminar-libro", Name: "Terminar un libro", XPPerUnit: 50, PointsPerUnit: 25},
	{CategorySlug: "lectura", Slug: "resumen-libro", Name: "Hacer resumen de un libro", XPPerUnit: 30, PointsPerUnit: 15},
	// Deporte
	{CategorySlug: "deporte", Slug: "ejercicio", Name: "Hacer ejercicio", XPPerUnit: 15, PointsPerUnit: 8},
	{CategorySlug: "deporte", Slug: "entrenamiento", Name: "Entrenamiento deportivo", XPPerUnit: 20, PointsPerUnit: 10},
	{CategorySlug: "deporte", Slug: "caminar", Name: "Salir a caminar", XPPerUnit: 8, PointsPerUnit: 4},
	// Arte
	{CategorySlug: "arte", Slug: "practicar-musica", Name: "Practicar instrumento", XPPerUnit: 15, PointsPerUnit: 8},
	{CategorySlug: "arte", Slug: "dibujar", Name: "Dibujar o pintar", XPPerUnit: 10, PointsPerUnit: 5},
	// Fe — vida espiritual. Eucaristía y Laudes pesan más en XP por su importancia.
	{CategorySlug: "fe", Slug: "eucaristia", Name: "Asistir a la Eucaristía", XPPerUnit: 50, PointsPerUnit: 25},
	{CategorySlug: "fe", Slug: "laudes", Name: "Rezar Laudes", XPPerUnit: 30, PointsPerUnit: 12},
	{CategorySlug: "fe", Slug: "rosario", Name: "Rezar el Rosario", XPPerUnit: 20, PointsPerUnit: 10},
	{CategorySlug: "fe", Slug: "lectura-biblia", Name: "Leer la Biblia", XPPerUnit: 12, PointsPerUnit: 6},
	{CategorySlug: "fe", Slug: "oracion-noche", Name: "Oración antes de dormir", XPPerUnit: 5, PointsPerUnit: 2},
	// Comidas — comer a tiempo. XP modesta pero la constancia es lo que importa.
	{CategorySlug: "comidas", Slug: "desayuno-rapido", Name: "Desayuno a tiempo", XPPerUnit: 5, PointsPerUnit: 2},
	{CategorySlug: "comidas", Slug: "almuerzo-rapido", Name: "Almuerzo a tiempo", XPPerUnit: 5, PointsPerUnit: 2},
	{CategorySlug: "comidas", Slug: "cena-rapida", Name: "Cena a tiempo", XPPerUnit: 5, PointsPerUnit: 2},
}

// Achievements — RPG-flavored Spanish milestones with humorous flair.
var Achievements = []AchievementSeed{
	{
		Slug: "quehaceres-aprendiz", Name: "Aprendiz Doméstico", Title: "Aprendiz de la Espuma",
		Description: "Has dado tus primeros pasos en las artes del hogar. El fregadero comienza a respetarte.",
		Combinator:  "ALL", BonusPoints: 20,
		Rules: []RuleSeed{{CategorySlug: "quehaceres", Metric: "count", Threshold: 10}},
	},
	{
		Slug: "quehaceres-veterano", Name: "Veterano del Hogar", Title: "Veterano de Mil Tareas",
		Description: "Cincuenta pequeños actos han forjado tu destino doméstico. Las escobas te saludan.",
		Combinator:  "ALL", BonusPoints: 50,
		Rules: []RuleSeed{{CategorySlug: "quehaceres", Metric: "count", Threshold: 50}},
	},
	{
		Slug: "quehaceres-antman", Name: "Antman del Hogar", Title: "El Antman del Nanoverso",
		Description: "Has limpiado tanto que Antman tropieza en el nanouniverso al contemplar tu obra.",
		Combinator:  "ALL", BonusPoints: 150,
		Rules: []RuleSeed{{CategorySlug: "quehaceres", Metric: "count", Threshold: 200}},
	},
	{
		Slug: "escuela-erudito", Name: "Erudito Estudiantil", Title: "El Erudito",
		Description: "Tu sed de conocimiento no tiene fin. Los libros te abren sus secretos.",
		Combinator:  "ALL", BonusPoints: 30,
		Rules: []RuleSeed{{CategorySlug: "escuela", Metric: "count", Threshold: 15}},
	},
	{
		Slug: "escuela-sabio", Name: "Sabio Académico", Title: "El Sabio del Aula",
		Description: "Quinientos puntos de experiencia académica. Tu sabiduría ilumina hasta los rincones más oscuros del aula.",
		Combinator:  "ALL", BonusPoints: 75,
		Rules: []RuleSeed{{CategorySlug: "escuela", Metric: "xp", Threshold: 500}},
	},
	{
		Slug: "lectura-explorador", Name: "Explorador de Mundos", Title: "Caminante entre Páginas",
		Description: "Cada libro es un mundo que has caminado. Las estanterías te miran con orgullo.",
		Combinator:  "ALL", BonusPoints: 40,
		Rules: []RuleSeed{{CategorySlug: "lectura", Metric: "count", Threshold: 5}},
	},
	{
		Slug: "lectura-bibliotecario", Name: "Bibliotecario Errante", Title: "Bibliotecario Errante",
		Description: "Veinte mundos cruzados. Las bibliotecas susurran tu nombre en sus pasillos.",
		Combinator:  "ALL", BonusPoints: 100,
		Rules: []RuleSeed{{CategorySlug: "lectura", Metric: "count", Threshold: 20}},
	},
	{
		Slug: "higiene-fragante", Name: "Fragante como Hada", Title: "Fragante",
		Description: "Hueles tan bien que las flores se inclinan a tu paso. Un perfume natural envidiable.",
		Combinator:  "ALL", BonusPoints: 15,
		Rules: []RuleSeed{{CategorySlug: "higiene", Metric: "count", Threshold: 30}},
	},
	{
		Slug: "deporte-atleta", Name: "Atleta en Forma", Title: "El Atleta",
		Description: "Tu cuerpo es tu templo y lo cuidas con devoción. La energía corre por tus venas.",
		Combinator:  "ALL", BonusPoints: 30,
		Rules: []RuleSeed{{CategorySlug: "deporte", Metric: "count", Threshold: 15}},
	},
	{
		Slug: "renacentista", Name: "Niño Renacentista", Title: "El Renacentista",
		Description: "Dominas las artes del hogar Y los misterios del saber. Da Vinci estaría orgulloso.",
		Combinator:  "ALL", BonusPoints: 60,
		Rules: []RuleSeed{
			{CategorySlug: "quehaceres", Metric: "count", Threshold: 20},
			{CategorySlug: "escuela", Metric: "count", Threshold: 15},
		},
	},
	{
		Slug: "equilibrado", Name: "Espíritu Equilibrado", Title: "El Equilibrado",
		Description: "Cuerpo, mente y casa: todo en armonía. Eres un hada del balance.",
		Combinator:  "ALL", BonusPoints: 100,
		Rules: []RuleSeed{
			{CategorySlug: "quehaceres", Metric: "count", Threshold: 20},
			{CategorySlug: "escuela", Metric: "count", Threshold: 15},
			{CategorySlug: "deporte", Metric: "count", Threshold: 10},
		},
	},
	{
		Slug: "versatil", Name: "Espíritu Versátil", Title: "El Versátil",
		Description: "Has tocado las artes, los libros o el deporte. Un generalista del crecimiento.",
		Combinator:  "ANY", BonusPoints: 50,
		Rules: []RuleSeed{
			{CategorySlug: "lectura", Metric: "count", Threshold: 10},
			{CategorySlug: "deporte", Metric: "count", Threshold: 10},
			{CategorySlug: "arte", Metric: "count", Threshold: 10},
		},
	},
	{
		Slug: "magnate", Name: "Magnate de los Puntos", Title: "El Magnate",
		Description: "Quinientos puntos en tu bolsillo. Una fortuna digna de leyenda.",
		Combinator:  "ALL", BonusPoints: 0,
		Rules: []RuleSeed{{CategorySlug: "", Metric: "points", Threshold: 500}},
	},
	{
		Slug: "super-magnate", Name: "Súper Magnate", Title: "El Súper Magnate",
		Description: "Mil puntos: dragones, princesas y videojuegos a tu alcance.",
		Combinator:  "ALL", BonusPoints: 0,
		Rules: []RuleSeed{{CategorySlug: "", Metric: "points", Threshold: 1000}},
	},
	{
		Slug: "caballero-hogar", Name: "Caballero del Hogar", Title: "Caballero",
		Description: "Has alcanzado el quinto nivel de tu maestría doméstica. La armadura de la escoba es tuya.",
		Combinator:  "ALL", BonusPoints: 40,
		Rules: []RuleSeed{{CategorySlug: "quehaceres", Metric: "level", Threshold: 5}},
	},
	{
		Slug: "senor-quehaceres", Name: "Señor de los Quehaceres", Title: "Señor de la Escoba",
		Description: "Décimo nivel doméstico. Las leyendas del hogar se cuentan tu nombre primero.",
		Combinator:  "ALL", BonusPoints: 200,
		Rules: []RuleSeed{{CategorySlug: "quehaceres", Metric: "level", Threshold: 10}},
	},

	// --- Lectura -------------------------------------------------------------
	{
		Slug: "lector-devoto", Name: "Lector Devoto", Title: "Devoto de las Páginas",
		Description: "Treinta sesiones con un libro abierto. Cada página es un peldaño más alto.",
		Combinator:  "ALL", BonusPoints: 60,
		Rules: []RuleSeed{{CategorySlug: "lectura", Metric: "count", Threshold: 30}},
	},
	{
		Slug: "cazador-mundos", Name: "Cazador de Mundos", Title: "Cazador de Mundos",
		Description: "Nivel 5 en Lectura — has caminado por mundos que pocos visitan.",
		Combinator:  "ALL", BonusPoints: 100,
		Rules: []RuleSeed{{CategorySlug: "lectura", Metric: "level", Threshold: 5}},
	},

	// --- Fe (vida espiritual) ------------------------------------------------
	{
		Slug: "aprendiz-alba", Name: "Aprendiz del Alba", Title: "Madrugador del Alba",
		Description: "Cuatro momentos de Fe. Tu día empieza con luz en el corazón.",
		Combinator:  "ALL", BonusPoints: 25,
		Rules: []RuleSeed{{CategorySlug: "fe", Metric: "count", Threshold: 4}},
	},
	{
		Slug: "peregrino-domingo", Name: "Peregrino del Domingo", Title: "Peregrino Dominical",
		Description: "200 XP en Fe — equivalente a varios domingos en Eucaristía. El camino se hace andando.",
		Combinator:  "ALL", BonusPoints: 75,
		Rules: []RuleSeed{{CategorySlug: "fe", Metric: "xp", Threshold: 200}},
	},
	{
		Slug: "hijo-luz", Name: "Hijo de la Luz", Title: "Hijo de la Luz",
		Description: "Treinta momentos sagrados. Tu lámpara arde sin descanso, como las vírgenes prudentes.",
		Combinator:  "ALL", BonusPoints: 75,
		Rules: []RuleSeed{{CategorySlug: "fe", Metric: "count", Threshold: 30}},
	},
	{
		Slug: "pequeno-discipulo", Name: "Pequeño Discípulo", Title: "Discípulo Fiel",
		Description: "500 XP en Fe. Has elegido seguir los pasos del Maestro con perseverancia.",
		Combinator:  "ALL", BonusPoints: 120,
		Rules: []RuleSeed{{CategorySlug: "fe", Metric: "xp", Threshold: 500}},
	},
	{
		Slug: "corazon-orante", Name: "Corazón Orante", Title: "Corazón en Oración",
		Description: "Nivel 5 en Fe — la oración es ya el latido de tu día.",
		Combinator:  "ALL", BonusPoints: 150,
		Rules: []RuleSeed{{CategorySlug: "fe", Metric: "level", Threshold: 5}},
	},

	// --- Comidas (puntualidad y disciplina en la mesa) -----------------------
	{
		Slug: "puntual-mesa", Name: "Puntual en la Mesa", Title: "Caballero de la Mesa",
		Description: "Diez comidas a tiempo. Tu silla no espera, te recibe.",
		Combinator:  "ALL", BonusPoints: 20,
		Rules: []RuleSeed{{CategorySlug: "comidas", Metric: "count", Threshold: 10}},
	},
	{
		Slug: "tres-tiempos", Name: "Maestro de los Tres Tiempos", Title: "Señor de Desayunos",
		Description: "Treinta comidas puntuales: el desayuno, el almuerzo y la cena conocen tu disciplina.",
		Combinator:  "ALL", BonusPoints: 50,
		Rules: []RuleSeed{{CategorySlug: "comidas", Metric: "count", Threshold: 30}},
	},
	{
		Slug: "disciplina-mesa", Name: "Disciplina de la Mesa", Title: "Disciplina de Hierro",
		Description: "Nivel 3 en Comidas. Llegar a tiempo es una pequeña victoria diaria.",
		Combinator:  "ALL", BonusPoints: 70,
		Rules: []RuleSeed{{CategorySlug: "comidas", Metric: "level", Threshold: 3}},
	},

	// --- Ahorro (puntos globales acumulados) ---------------------------------
	// Estos son escalones progresivos que celebran ahorrar antes de gastar.
	{
		Slug: "ahorrador-novato", Name: "Ahorrador Novato", Title: "Aprendiz del Cofre",
		Description: "Tus primeras 100 monedas. La paciencia es un escudo.",
		Combinator:  "ALL", BonusPoints: 0,
		Rules: []RuleSeed{{CategorySlug: "", Metric: "points", Threshold: 100}},
	},
	{
		Slug: "ahorrador-cuidadoso", Name: "Ahorrador Cuidadoso", Title: "Guardián del Cofre",
		Description: "250 monedas en tu bolsillo. Hoy posees algo que muchos no llegan a ver.",
		Combinator:  "ALL", BonusPoints: 0,
		Rules: []RuleSeed{{CategorySlug: "", Metric: "points", Threshold: 250}},
	},
	{
		Slug: "tesorero-leyenda", Name: "Tesorero Legendario", Title: "Tesorero Legendario",
		Description: "2500 monedas. Tu cofre rebosa y los bardos cantan tu nombre.",
		Combinator:  "ALL", BonusPoints: 0,
		Rules: []RuleSeed{{CategorySlug: "", Metric: "points", Threshold: 2500}},
	},
	{
		Slug: "gran-tesorero", Name: "Gran Tesorero", Title: "Gran Tesorero del Reino",
		Description: "5000 monedas — solo los más constantes llegan a este peldaño.",
		Combinator:  "ALL", BonusPoints: 0,
		Rules: []RuleSeed{{CategorySlug: "", Metric: "points", Threshold: 5000}},
	},
}
