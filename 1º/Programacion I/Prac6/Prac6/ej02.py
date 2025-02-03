name_file = input('Introduce el nombre de un fichero: ')

file = open(name_file, 'r')
lineas = file.readlines()

correcion = lineas[0]

for linea in lineas[1:]:
    dni, respuesta = linea.strip().split('#')
    print(f'El alumno {dni} a dado las respuestas {respuesta}')

print(f'Las respuestas correctas son {correcion}')

file.close()