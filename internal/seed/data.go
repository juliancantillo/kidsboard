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
var Categories = []CategorySeed{
	{Slug: "quehaceres", Name: "Quehaceres", Description: "Tareas del hogar y responsabilidades.", Icon: "🧹", Color: "#10B981"},
	{Slug: "escuela", Name: "Escuela", Description: "Tareas escolares y aprendizaje.", Icon: "📚", Color: "#3B82F6"},
	{Slug: "higiene", Name: "Higiene", Description: "Cuidado personal y aseo.", Icon: "🧼", Color: "#06B6D4"},
	{Slug: "lectura", Name: "Lectura", Description: "Hábitos de lectura y exploración de historias.", Icon: "📖", Color: "#8B5CF6"},
	{Slug: "deporte", Name: "Deporte", Description: "Actividad física y deportes.", Icon: "⚽", Color: "#F97316"},
	{Slug: "arte", Name: "Arte y Música", Description: "Creatividad, música y expresión artística.", Icon: "🎨", Color: "#EC4899"},
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
	{CategorySlug: "lectura", Slug: "terminar-libro", Name: "Terminar un libro", XPPerUnit: 50, PointsPerUnit: 25},
	// Deporte
	{CategorySlug: "deporte", Slug: "ejercicio", Name: "Hacer ejercicio", XPPerUnit: 15, PointsPerUnit: 8},
	{CategorySlug: "deporte", Slug: "entrenamiento", Name: "Entrenamiento deportivo", XPPerUnit: 20, PointsPerUnit: 10},
	{CategorySlug: "deporte", Slug: "caminar", Name: "Salir a caminar", XPPerUnit: 8, PointsPerUnit: 4},
	// Arte
	{CategorySlug: "arte", Slug: "practicar-musica", Name: "Practicar instrumento", XPPerUnit: 15, PointsPerUnit: 8},
	{CategorySlug: "arte", Slug: "dibujar", Name: "Dibujar o pintar", XPPerUnit: 10, PointsPerUnit: 5},
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
}
