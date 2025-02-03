""" Disponemos de una serie de ficheros de texto con informaci´on sobre donaciones de sangre en una determinada
provincia y a˜no. Cada l´ınea de uno de estos ficheros contiene, separados por el car´acter almohadilla (#), los datos
de una donaci´on: punto de recogida, DNI del donante, grupo sangu´ıneo y mes (n´umero entero de 1 a 12) en que
se realiz´o. Por ejemplo, el contenido del fichero castell´on2022.txt podr´ıa ser as´ı:

ESTCE#012345678X#A+#7
C.C. Salera#876543210J#AB-#3
Hospital Provincial#55544433Z#0-#3
ESTCE#012345678#A+#12
C.C. Salera#33333333A#A+#7

Implementa en Python:

a) (2 puntos) Una funci´on, donaciones_por_mes, que tenga como par´ametros el nombre de un fichero de dona-
ciones (fichero de texto con el formato descrito antes) y un grupo sangu´ıneo. La funci´on debe crear y devolver
una lista de 12 valores enteros que almacene la cantidad de donaciones del grupo sangu´ıneo indicado que se
han producido en cada mes del a˜no. Solo se puede recorrer el fichero dado una vez. Por ejemplo, la llamada
donaciones_por_mes(’castell´on2022.txt’,’A+’) devolver´ıa [0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1]. """

def donaciones_por_mes(nombre_fichero, grupo_sanguineo):
    lista_gruposanguineo = [0]*12

    with open(nombre_fichero, 'r') as file:
        for linea in file:
            partes = linea.strip().split('#')
            hospital = partes[0]
            dni = partes[1]
            grupo = partes[3]
            mes = int(partes[4])

            if grupo == grupo_sanguineo:    
                for i in range(len(lista_gruposanguineo)):
                    if mes - 1 == i: 
                        lista_gruposanguineo[i] += 1

        return lista_gruposanguineo
    
""" b) (2 puntos) Una funci´on, mayor_periodo_creciente, que tenga como par´ametro una lista como la construida
en el apartado anterior y devuelva el mayor n´umero de meses consecutivos en los que la cantidad de donaciones
ha aumentado con respecto al mes anterior, o cero si no ha habido ning´un aumento. Por ejemplo, la llama-
da mayor_periodo_creciente([2, 4, 2, 5, 3, 1, 8, 2, 5, 6, 4, 2]) devolver´ıa 3 por la longitud de la
secuencia 2, 5, 6 """

def mayor_periodo_creciente(lista_gruposanguineo):
    tramo_max = 0
    contador =  1
    
    for i in range(1, len(lista_gruposanguineo)):
        if lista_gruposanguineo[i] > lista_gruposanguineo[i-1]:
            contador += 1
        else:
            if contador > tramo_max:
                tramo_max = contador
            contador = 1

    if contador > tramo_max:
        tramo_max = contador

    return tramo_max

""" c) (0,5 puntos) Una funci´on, buscar_pto_recogida, que tenga como par´ametros una lista de objetos de la clase
Donaciones (mostrada a continuaci´on) y un punto de recogida. La funci´on debe buscar en la lista un objeto
cuyo atributo pto_recogida coincida con el punto de recogida recibido y devolver dicho objeto. Si no lo
encuentra, debe devolver None """

class Donaciones:
    def __init__(self, pto_recogida):
        self.pto_recogida = pto_recogida
        self.cantidad = 0
    def añadir(self):
        self.cantidad += 1

def buscar_pto_recogida(lista, pto_recogida):
    for donacion in lista:
        if donacion.pto_recogida == pto_recogida:
            return pto_recogida
    return None

""" d) (1,5 puntos) Una funci´on, donaciones_por_punto_recogida, que tenga como par´ametro el nombre de un
fichero de donaciones. La funci´on debe crear y devolver una lista de objetos de la clase Donaciones. La lista
debe almacenar un ´unico objeto por cada punto de recogida con el n´umero total de donaciones recogidas all´ı.
Esta funci´on debe usar la funci´on buscar_pto_recogida del apartado anterior. Solo se puede recorrer el fichero
dado una vez. Por ejemplo, la llamada donaciones_por_punto_recogida(’castell´on2022.txt’) devolver´ıa
la lista siguiente: """

def donaciones_por_punto_recogida(nombre_fichero):
    lista_recogidas = []

    with open(nombre_fichero, 'r') as file:
        for linea in file:
            partes = linea.strip().split('#')
            pto_recogida = partes[0]

            donacion_obj = buscar_pto_recogida(lista_recogidas, pto_recogida)

            if donacion_obj:
                donacion_obj.cantidad += 1
            else:
                nueva_donacion = Donaciones(pto_recogida)
                nueva_donacion.cantidad = 1
                lista_recogidas.append(nueva_donacion)
        
        return lista_recogidas