""" Tenemos una serie de ficheros de texto con informaci´on acerca de los deportistas de alto rendimiento que actual-
mente cursan estudios en diversas universidades espa˜nolas. Cada l´ınea de uno de estos ficheros (correspondiente a
una universidad) contiene, separados por el car´acter #, el DNI del deportista, la modalidad deportiva que practica
y la titulaci´on en que est´a matriculado. A continuaci´on se muestran, a modo de ejemplo, algunas l´ıneas de un
fichero como el descrito:

12345678A#Atletismo#Ingenier´ıa Inform´atica
24681357B#Rugby#Medicina
33468648C#Ciclismo#Matem´atica Computacional
56523475D#Voleibol#Derecho
65234523E#Ciclismo#Ingenier´ıa Inform´atica
Considera la clase Deporte, definida como sigue:
class Deporte:

def __init__(self, nombre):
self.nombre = nombre # Modalidad deportiva
self.estudiantes = [] # DNI de los estudiantes que practican esa modalidad deportiva

Implementa en Python:

a) (2 puntos) Una funci´on, crear_lista_deportes, que tenga como par´ametros el nombre de un fichero de
texto como el descrito anteriormente y el nombre de una titulaci´on. La funci´on debe crear y devolver una
lista de objetos de la clase Deporte, de forma que cada modalidad deportiva practicada por estudiantes de esa
titulaci´on aparezca una ´unica vez en la lista. El atributo estudiantes de cada objeto debe contener los DNI
de todos los deportistas de esa titulaci´on que practican esa modalidad.
Nota: Solo se permite recorrer el fichero una vez. """

class Deporte:
    def __init__(self, nombre):
        self.nombre = nombre
        self.estudiantes = []
    
def crear_lista_deportes(nombre_fichero, nombre_titulacion):
    lista_deportes = []

    with open(nombre_fichero, 'r') as file:
        for linea in file:
            partes = linea.strip().split('#')
            DNI = partes[0]
            deporte = partes[1]
            titulacion = partes[2]

            
            if titulacion == nombre_titulacion:

                modalidad_existente = None
                for modalidad in lista_deportes:
                    if modalidad.nombre == deporte:
                        modalidad_existente = modalidad
                        break
                    else:
                        nueva_modalidad = Deporte(deporte)
                        nueva_modalidad.estudiantes.apppend(DNI)
                        lista_deportes.append(nueva_modalidad)

        return lista_deportes
    
""" b) (1 punto) Una funci´on, modalidad_m´as_representada, que tenga como par´ametro una lista de objetos como
la construida en el apartado anterior y devuelva el nombre de la modalidad deportiva con mayor n´umero de
estudiantes de alto rendimiento. En caso de empate entre varias modalidades deportivas, la funci´on puede
devolver cualquiera de ellas """

def modalidad_mas_representada(lista_deportes):
    modalidad_max = None
    max_estudiantes  = 0
    
    for deporte in lista_deportes:
        if len(deporte.estudiandes) > max_estudiantes:
            max_estudiantes = len(deporte.estudiantes)
            modalidad_max = deporte.nombre

    return modalidad_max

""" c) (2 puntos) Una funci´on, hay_representantes, que tenga como par´ametros una lista de objetos como la
construida en el primer apartado y otra lista de cadenas con nombres de modalidades deportivas. La funci´on
debe devolver True cuando la titulaci´on cuente con representantes para todas las modalidades deportivas
indicadas en la lista dada y False en caso contrario """

def hay_representantes(lista_deportes, modalidades_deportivas):
    for modalidad in modalidades_deportivas:
        encontrado = False
        for deporte in lista_deportes:
            if deporte.nombre == modalidad:
                encontrado = True
                break

        if not encontrado:
            return False
    
    return True