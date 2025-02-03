""" a) (2,5 puntos) Una funci´on, crear_matriz_presencia_al´ergenos, que tenga como par´ametros el nombre de
un fichero de texto con el formato descrito anteriormente y una lista con los nombres de todos los al´ergenos de
obligada declaraci´on1. La funci´on debe crear y devolver una matriz de M × N valores l´ogicos, donde M es el
n´umero de platos en la carta del restaurante y N el n´umero de al´ergenos en la lista dada. Cada celda [i][j]
de la matriz debe almacenar True o False seg´un el plato i de la carta contenga o no el al´ergeno j.
Por ejemplo, considerando el fichero de texto mostrado anteriormente y la siguiente lista de al´ergenos:
[’Cereales con gluten’, ’Huevos’, ’L´acteos’, ’Frutos secos’, ’Pescado’, ’Sulfitos’, ...]
la funci´on deber´ıa devolver la siguiente matriz de presencia de al´ergenos:
[
[False, True, True, True, False, True, ... ], # Ensalada templada de esp´arragos trigueros
[False, False, False, False, False, False, ... ], # Menestra de verduras con jam´on ib´erico
[True, False, False, False, True, True, ... ], # Lubina al horno con verduritas confitadas
[False, False, False, False, True, False, ... ], # Bacalao a la parrilla con cebollitas, tomate...
[False, False, True, False, False, True, ... ], # Magret de pato con pur´e de pera y queso cremoso
[False, False, False, False, False, False, ... ], # Cordero lechal asado
[False, False, True, False, False, False, ... ], # Queso de oveja zamorano
... # Resto de platos
] """

def crear_matriz_presencia_alergenos(nombre_fichero, lista_alergenos_obligada_declaracion):
    matriz = []
    
    with open(nombre_fichero, 'r') as file:
        for linea in file:
            partes = linea.strip().split(':')
            platos = partes[0]
            alergenos = partes[1].strip().split(',')

            fila = []
            for alergeno in lista_alergenos_obligada_declaracion:
                fila.append(alergeno in alergenos)

            matriz.append(fila)

    return matriz

""" b) (1,5 puntos) Una funci´on, al´ergeno_m´as_frecuente, que tenga como par´ametros la lista de al´ergenos de
obligada declaraci´on y una matriz de presencia de al´ergenos y devuelva el nombre del al´ergeno m´as frecuente
en la matriz dada. En caso de empate entre varios al´ergenos, puedes devolver cualquiera de ellos """

def alergeno_mas_frecuente(lista_alergenos_declaracion_obligatoria, matriz):
    frecuencia = [0] * len(lista_alergenos_declaracion_obligatoria)
    
    for j in range(len(matriz[0])):
        for i in range(len(matriz)):
            if matriz[i][j] == True:
                frecuencia += 1

    mas_repetido = -1
    indice_mas_repetido = -1

    for indice in range(len(frecuencia)):
        if frecuencia[indice] > mas_repetido:
            mas_repetido = frecuencia[indice]
            indice_mas_repetido = indice

    return lista_alergenos_declaracion_obligatoria[indice_mas_repetido]

""" c) (2 puntos) Una funci´on, cantidad_platos_aptos, que tenga como par´ametros la lista de al´ergenos de obligada
declaraci´on, una matriz de presencia de al´ergenos y la lista de al´ergenos no tolerados por un cliente. La funci´on
debe devolver la cantidad de platos de la carta aptos para el cliente, es decir, que no contienen ninguno de los
al´ergenos indicados """

def cantidad_platos_aptos(lista_alergenos_declaración_obligatoria, matriz, alergenos_no_tolerados):
    platos_aptos = 0

    for fila in matriz:
        apto = True
        for j in range(len(lista_alergenos_declaración_obligatoria)):
            if lista_alergenos_declaración_obligatoria[j] in alergenos_no_tolerados and fila[j] == True:
                apto = False
                break

        if apto == True:
            platos_aptos += 1
    
    return platos_aptos