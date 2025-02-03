#Ver los numeros abundantes menores al numero introducido
num = int(input('Introduce un numero: '))

for divisores in range (2, num):    #Mirar los numeros del numerador
    suma_divisores = 0
    for i in range (1, divisores):  #Mirar los numeros del denominador
        if divisores % i  == 0:     #Observa que denominadores son = 0 y posteriormente los suma
            suma_divisores += i
        
    if suma_divisores > divisores:  #Si la suma de dichos denominadores es mayor al numerador, ese numerador será abundante menor al introducido
        print(divisores)