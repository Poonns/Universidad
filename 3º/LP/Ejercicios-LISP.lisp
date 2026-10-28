;Sin definir otras funciones adicionales, implementa una función cuenta_exitos que reciba una función de evaluación (que devuelve T o NIL) y un número arbitrario de argumentos. Debe devolver un único número indicando cuántos elementos cumplen dicha condición.
;(cuenta_exitos #'evenp 1 2 3 4 5 6) -> 3
;Tu objetivo: Resolverlo en una sola línea usando tu combo de mapcar.
(defun cuenta_exitos(f &rest a)
    (length(remove NIL(mapcar f a)))
)

;Sin definir otras funciones adicionales, implementa una función emparejar_con_fijo que, dado un elemento fijo y un número arbitrario de argumentos adicionales, devuelva una lista de listas donde cada argumento esté emparejado con el elemento fijo en la primera posición.
;(emparejar_con_fijo 'X 1 2 3) -> ((X 1) (X 2) (X 3))
;Tu objetivo: Resolverlo de forma estrictamente recursiva (sin mapcar). Pon mucha atención a dónde usas cons y dónde usas list para no anidar de más.
(defun emparejar_con_fijo(e &rest a)
    (cond  
        ((null a)nil)
        (t
            (cons(list e (car a)(apply #'emparejar_con_fijo e (cdr a))))
        )
    )
)

;Sin definir otras funciones adicionales ni usar variables locales (let), implementa una función separar_pares_impares que reciba un número arbitrario de enteros y devuelva una lista con dos sublistas: la primera con los números pares (evenp) y la segunda con los impares (oddp).
;(separar_pares_impares 1 2 3 4 5) -> ((2 4) (1 3 5))
;(separar_pares_impares 2 4) -> ((2 4) NIL)
(defun separar_pares_impares(&rest a)
    (cond
        ((null a)(list nil nil))
        ((zerop (car a))
            (apply #'separar_pares_impares (cdr a))
        )
        ((if(=(mod 2 (car a))0))
            (list
                (cons(car a)(car(apply #'separar_pares_impares (cdr a))))
                (cadr(apply #'separar_pares_impares (cdr a)))
            )
        )
        (t
            (list
                (car(apply #'separar_pares_impares (cdr a)))
                (cons(car a)(cadr(apply #'separar_pares_impares (cdr a))))
            )
        )
    )
)
,
;Reto 1: La Matriz Cruzada (Anidamiento Funcional vs Recursividad de Doble Capa)
;Sin definir otras funciones adicionales, implementa multiplicadores_locos que reciba una lista de números "base" y un número arbitrario de multiplicadores sueltos. Debe devolver una lista de listas, donde cada número base se haya multiplicado por todos los multiplicadores.
;(multiplicadores_locos '(1 2 3) 10 20) -> ((10 20) (20 40) (30 60))
;(multiplicadores_locos '(5) 2 3 4) -> ((10 15 20))

;Version mapcar
(defun multiplicadores_locos (l &rest m)
    (mapcar(lambda(x)((list(mapcar(lambda(multiplicador)(* multiplicador x))m))))l)
)

;Version recursiva
(defun multiplicadores_locos (l &rest m)
    (cond
        ((null m)l)
        (t
            (cons(mapcar(lambda(x)(* (car l) x)))(apply #'multiplicadores_locos (cdr l) m))
        )
    )
)

;Sin definir otras funciones adicionales ni usar variables locales (let), implementa dividir_transformar que reciba un número arbitrario de enteros. Debe devolver una lista con dos sublistas: la primera contendrá los números pares multiplicados por 10, y la segunda contendrá los números impares restándoles 1 (puedes usar la primitiva 1-).
;(dividir_transformar 1 2 3 4 5) -> ((20 40) (0 2 4))

;Version mapcar
(defun dividir_transformar (&rest enteros)
    (list
        (remove nil(mapcar (lambda (x)(if(=(mod x 2)0)(* x 10) nil)) enteros))
        (remove nil (mapcar (lambda (x)(if(=(mod x 2)1)(1- x) nil)) enteros))
    )
)

;Version recursiva
(defun dividir_transforma (&rest enteros)
    (cond
        ((null enteros)(list nil nil))
        ((= (mod (car enteros) 2)0)
            (list
                (cons (* 10 (car enteros)) (car(apply #'dividir_transforma (cdr enteros))))
                                   (cadr(apply #'dividir_transforma (cdr enteros)))
            )
        )
        (t
            (list
                                        (car(apply #'dividir_transforma (cdr enteros)))
                (cons (1- (car enteros))(cadr(apply #'dividir_transforma (cdr enteros))))
            )
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplica_si que reciba una función de condición (predicado), una función de transformación y un número arbitrario de listas. Debe devolver una lista con el mismo número de sublistas, donde a los elementos individuales que cumplan la condición se les haya aplicado la transformación, y los que no, se queden exactamente igual.
;(aplica_si #'evenp (lambda (x) (* x 10)) '(1 2 3) '(4 5) '(6)) → ((1 20 3) (40 5) (60))
(defun aplica_si (f_cond f_trans &rest l)
    (mapcar(lambda(sublista)(mapcar(lambda(y)(if(funcall f_cond y)(funcall f_trans y) y))sublista))l)
)

;Siguiendo la lógica del ejercicio contarCondiciones de la segunda convocatoria, pero un paso más allá. Sin definir otras funciones adicionales, implementar una función en lisp llamada clasificar_valores que, dada una lista sin anidamientos de valores y un número arbitrario de funciones que representen condiciones, devuelva una lista de listas. Cada sublista debe contener los valores exactos de la lista original que cumplen con la condición correspondiente.
;(clasificar_valores '(1 2 3 4) #'evenp #'oddp (lambda (x) (> x 2))) → ((2 4) (1 3) (3 4))
(defun clasisicar_valores (l &rest f_cond)
    (cond
        ((null f_cond)(list nil))
        (t
            (cons(remove nil(mapcar(lambda(elemento)(if(funcall (car f_cond) elemento)elemento nil))l)))
                    (apply #'clasificar_valores l (cdr f_cond))
        )
    )
)

;Tomando la idea del ejercicio primeroCentralyUltimo. Sin definir otras funciones adicionales, implementar una función en lisp llamada extraer_extremos que, dado un número arbitrario de listas suministradas como parámetros, devuelva una lista de listas donde cada sublista contenga exactamente dos elementos: el primer y el último elemento de la lista original. Si la lista tiene un solo elemento, este se repetirá. Si está vacía, devolverá (NIL NIL).
;(extraer_extremos '(1 2 3) '(9) '() '(5 6 7 8)) → ((1 3) (9 9) (NIL NIL) (5 8))

;Version mapcar
(defun extraer_extremos (&rest l)
    (mapcar(lambda(sublista)(list (car sublista)(car(reverse sublista))))l)
)

;Version recursiva
(defun extraer_extremos (&rest l)
    (cond
        ((null l)(list nil nil))
        (t
            (cons(list
                (car(car l))(car(reverse(car l)))
                    (apply #'extraer_extremos (cdr l))
            ))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada procesar_mixto que reciba una función unaria y un número arbitrario de argumentos que pueden ser números o listas. La función debe devolver una nueva lista plana donde se haya aplicado la función suministrada directamente sobre el argumento (si este es un número), o sobre la longitud de la lista (si el argumento es una lista).  
;(procesar_mixto #'1+ 5 '(A B C) 10 NIL '(X Y)) -> (6 4 11 1 3)

;Version mapcar
(defun procesar_mixto (f &rest a)
    (mapcar(lambda(x)(if(listp x)(funcall f (length x))(funcall f x)))a)
)

;Version recursiva
(defun procesar_mixto (f &rest a)
    (cond
        ((null a) nil)
        ((listp (car a))
            (cons(funcall f (length (car a)))(apply #'procesar_mixto f (cdr a)))
        )
        (t
            (cons(funcall f (car a))(apply #'procesar_mixto f (cdr a)))
        )
    )
)   

;Sin definir otras funciones adicionales y sin usar llamadas recursivas, implementar una función en LISP llamada operar_extremos que, dada una función binaria (como #'+ o #'*) y un número arbitrario de listas no vacías, devuelva una lista plana. Cada elemento de esta lista devuelta será el resultado de aplicar la función binaria entre el primer y el último elemento de la sublista correspondiente.
;(operar_extremos #'+ '(1 2 3) '(10 20) '(5 5 5 5)) -> (4 30 10) 
(defun operar_extremos (f_bin &rest l)
    (mapcar(lambda(sublista)(funcall f_bin (car sublista) (car(reverse sublista))))l)
)

;Sin definir otras funciones adicionales y sin utilizar la función mapcar, implementar una función recursiva en LISP llamada filtrar_tipos que, dado un predicado (una función que devuelve T o NIL, como #'numberp o #'listp) y un número arbitrario de argumentos mezclados, devuelva una única lista plana que contenga exclusivamente los argumentos que cumplen dicho predicado.
;(filtrar_tipos #'numberp 1 '(A) 2 NIL 3) -> (1 2 3)
(defun filtrar_tipos (f_pred &rest a)
    (cond
        ((null a) nil)
        ((funcall f_pred (car a))
            (cons(car a)(apply #'filtrar_tipos f_pred (cdr a)))
        )
        (t
            (apply #'filtrar_tipos f_pred (cdr a))
        )
    )
)

;Sin definir otras funciones adicionales y utilizando únicamente mapcar (sin recursividad), implementar una función en LISP llamada cruzar_funciones que, dado un valor cualquiera y un número arbitrario de funciones, devuelva una lista de pares. Cada "par" será una sublista de dos elementos: el valor original y el resultado de aplicarle la función correspondiente.
;(cruzar_funciones 5 #'1+ #'1- (lambda (x) (* x 2))) -> ((5 6) (5 4) (5 10))
(defun cruzar_funciones (n &rest f)
    (mapcar(lambda(funcion)(list n (funcall funcion n)))f)
)

;Sin definir otras funciones adicionales y sin utilizar la función mapcar, implementar una función puramente recursiva en LISP llamada suma_estricta que, dado un número arbitrario de argumentos de distintos tipos (números, listas, átomos nulos), devuelva la suma matemática exclusivamente de los argumentos que sean números, ignorando por completo el resto. En caso de no haber ningún número, debe devolver 0.
;(suma_estricta 10 '(A B) 5 NIL 20) -> 35
;(suma_estricta '(1 2) NIL) -> 0
(defun suma_estricta (&rest a)
    (cond
        ((null a) 0)
        ((numberp (car a))
            (+ (car a)(apply #'suma_estricta (cdr a)))
        )
        (t
            (apply #'suma_estricta (cdr a))
        )
    )
)

;Sin definir otras funciones y usando exclusivamente recursividad pura (prohibido mapcar, flatten o similares), implementa la función sumar_profundidad.
;Esta función recibe una función de predicado (ej. #'evenp, #'oddp) y una lista que puede contener números, símbolos, o cualquier cantidad de listas anidadas a cualquier nivel. Debe devolver la suma matemática de todos los números en cualquier nivel de profundidad que cumplan el predicado. Si no hay ninguno, devuelve 0.
;(sumar_profundidad #'evenp '(1 (2 A) ((3 4) NIL 5) B (6))) -> 12  ;; (Suma de 2 + 4 + 6)
(defun sumar_profundidad (f_pred l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (apply #'sumar_profundidad f_pred (car l))(apply #'sumar_profundidad f_pred (cdr l))
        )
        (t
            (if((funcall f_pred (car l))
                (+(car l)(apply #'sumar_profundidad f_pred (cdr l))))
                (apply #'sumar_profundidad f_pred (cdr l)))
        )
    )
)

;Sin usar recursividad y utilizando únicamente mapcar, implementa la función batalla_funciones.
;Recibe tres argumentos: una lista de valores numéricos, una lista de predicados, y una lista de funciones transformadoras. (Asume que las tres listas tienen la misma longitud).
;La función debe evaluar cada valor con su predicado correspondiente en la misma posición. Si el predicado devuelve T, se le aplica la transformación correspondiente. Si devuelve NIL, el valor se queda igual. Devuelve la lista resultante.
;(batalla_funciones '(2 3 4) '(#'evenp #'evenp #'oddp) '(#'1+ #'1- (lambda (x) (* x 10)))) -> (3 3 4)
(defun batalla_funciones (l_val l_pred l_func)
    (mapcar (lambda (v p f)(if (funcall p v))(funcall f v) v) l_val l_pred l_func)
)

;Sin usar llamadas recursivas y usando puramente lógica funcional (mapcar, remove, apply...), implementa la función limpieza_estructural.
;Recibe un predicado y un número arbitrario de listas mixtas (pueden contener números y otras cosas). Debe devolver una lista de listas donde, de cada sublista, se hayan eliminado los elementos que NO cumplen el predicado.
;La trampa mortal: Si tras el filtrado una sublista queda completamente vacía, NO debe aparecer en el resultado final.
;(limpieza_estructural #'numberp '(A 1 B 2) '(X Y Z) '(3 NIL 4)) -> ((1 2) (3 4)) 
(defun limpieza_estructural (f &rest l)
    (remove nil(mapcar (lambda(sublista)(remove nil(mapcar (lambda(x)(if(funcall f x)x nil))sublista))l)))
)

;Basado en tus ejercicios aplanar y filtrar, vamos a dar un paso más allá.
;Implementa una función llamada (recolectar_si predicado arbol) que reciba una función unaria que devuelva un booleano (predicado) y una lista que puede tener cualquier nivel de anidamiento (arbol). La función debe devolver una lista plana (sin anidar) con todos los átomos del árbol para los cuales la función predicado devuelva T.
;(recolectar_si #'evenp '(1 (2 3 (4)) 5 (6 (7 8)))) ->(2 4 6 8)
(defun recolectar_si (f l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (append (recolectar_si f (car l))(recolectar_si f (cdr l)))
        )
        ((funcall f (car l))
            (cons (car l)(recolectar_si f (cdr l)))
        )
        (t
            (recolectar_si f (cdr l))
        )
    )
)

;Implementa una función llamada (pipeline x &rest funciones) que reciba un valor inicial x y un número arbitrario de funciones unarias. La función debe aplicar las funciones en cadena de izquierda a derecha. Es decir, el resultado de la primera función se le pasa a la segunda, y así sucesivamente: fn(...f2(f1(x))...)fn(...f 2(f 1​	(x))...). Si no se le pasa ninguna función, debe devolver x
;(pipeline 3 (lambda (x) (+ x 2)) (lambda (x) (* x 2)) #'1-) -> 9   ;; Explicación: 3+2 = 5 -> 5*2 = 10 -> 10-1 = 9
(defun pepeline (x &rest funciones)
    (cond
        ((null funciones) x)
        (t
            (apply #'pepeline (funcall (car funciones) x) (cdr funciones))
        )
    )
)

;Implementa ahora la función (diagonal M) que, dada una matriz cuadrada M, devuelva una lista plana con los elementos de su diagonal principal.
;(Pista: Seguramente necesites una función auxiliar que lleve cuenta de la posición, o jugar con la recursividad extrayendo el car de la primera fila y avanzando en las siguientes).
(defun diagonal (m)
    (cond
        ((null m) nil)
        (t
            (cons (caar m)(diagonal(mapcar #'cdr (cdr m))))  
        )
    )
)

;Implementa una función (producto_cartesiano A B) que reciba dos listas planas (que representan conjuntos) y devuelva una lista de listas con todas las parejas posibles formadas por un elemento de A y un elemento de B.
(defun producto_cartesiano (a b)
   (apply #append(mapcar (lambda (x)(mapcar (lambda (y)((list x y))b)))a))
)

;Implementa una función (duplicar_elementos lista) que reciba una lista plana y devuelva otra lista donde cada elemento aparezca dos veces seguidas.
(defun duplicar_elementos (l)
    (apply #append (mapcar (lambda(x)(list x x))l))
)

;Implementa tu propia versión de la función nativa reverse, llamándola (mi_reverso lista). Debe coger una lista y devolverla dada la vuelta usando recursividad pura.
(defun mi_reverso (l)
    (cond
        ((null l) nil)
        (t
            (append (mi_reverso(cdr l)(list (car l))))
        )
    )
)

;Implementa una función (sustituir_si predicado nuevo_valor arbol) que reciba una función unaria (predicado), un valor y una lista con cualquier nivel de anidamiento (arbol). La función debe devolver el mismo árbol intacto en su estructura, pero sustituyendo por el nuevo_valor todos los átomos para los cuales la función predicado devuelva T.
(defun sustituir_si (f v l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (sustituir_si f v (car l))(sustituir_si f v (cdr l)))
        )
        ((funcall f (car l))
            (cons v (sustituir_si f v (cdr l)))
        )
        (t
            (cons (car l)(sustituir_si f v (cdr l)))
        )
    )
)

;Implementa una función (multi_mapcar lista &rest funciones) que reciba una lista plana y un número arbitrario de funciones unarias. La función debe devolver una lista de listas, donde cada sublista contenga el resultado de aplicar todas y cada una de las funciones al elemento correspondiente de la lista original.
(defun multi_mapcar (l &rest f)
    (mapcar (lambda(x)(mapcar (lambda(y)(funcall y x))f))l)
)

;Implementa una función (sumar_columnas matriz) que reciba una matriz (lista de listas) y devuelva una lista plana con la suma matemática de los elementos de cada columna. Prohibido usar funciones auxiliares; resuélvelo mediante recursividad o explotando mapcar y apply.
(defun sumar_columnas (m)
    (apply #'mapcar #'+ m)
)

;Implementa la función (comprimir lista) que reciba una lista plana y devuelva una lista de listas, donde cada sublista contenga el elemento y el número de veces consecutivas que aparece en la lista original.
(defun comprimir (l)
    (cond
        ((null l) nil)
        ((= (car l) (caar(comprimir(cdr l))))
            (cons(list (car l (1+ (cadar)(comprimir (cdr l)))))
                    (comprimir (cdr l)))
        )
        (t
            (cons (list (car l) 1)(comprmir (cdr l)))
        )
    )
)

;Implementa una función (map_arbol f arbol) que reciba una función unaria f y una lista con cualquier nivel de anidamiento arbol. La función debe devolver un árbol con la misma estructura exacta, pero con la función aplicada a cada uno de sus átomos.
(defun map_arbol (f l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (map_arbol f (car l))(map_arbol f (cdr l)))
        )
        (t
            (cons (funcall f (car l))(map_arbol f (cdr l)))
        )
    )
)

;Basado en el Ejercicio 2 de tu Práctica 1. Implementa una función (aplicar_alterno f g &rest valores) que reciba dos funciones unarias y una cantidad variable de valores. Debe devolver una lista plana aplicando f al primer valor, g al segundo, f al tercero, g al cuarto, y así sucesivamente.
(defun aplicar_alterno (f g &rest valores)
    (cond
        ((null valores) nil)
        (t
            (cons (funcall f (car valores))(apply #'aplicar_alterno g f (cdr valores)))
        )
    )
)

;En tu Práctica 1 te piden implementar la matriz transpuesta. Una matriz transpuesta convierte las filas en columnas y viceversa. Implementa (transpuesta m) sin usar recursividad explícita
(defun traspuesta (m)
    (apply #'mapcar #'list m)
)

;Implementa la función (interseccion A B) que reciba dos listas planas que representan conjuntos, y devuelva una lista plana solo con los elementos de A que también estén presentes en B. Prohibido usar funciones auxiliares o la función nativa de intersección.
(defun interseccion (a b)
    (cond
        ((OR (null a)(null b)) nil)
        ((member (car a) b)
            (list (car a)(interseccion (cdr a) b))
        )
        (t
            (interseccion (cdr a) b)
        )
    )
)

;Sin utilizar funciones de orden superior (mapcar, remove, etc.) y usando exclusivamente recursividad pura, implementa la función aplanar_condicional.
;La función recibe un predicado y una lista de profundidad arbitraria (un árbol que puede contener átomos, NIL y listas dentro de listas). Debe devolver una lista plana (un solo nivel) que contenga únicamente los átomos de toda la estructura que cumplan el predicado.
;(aplanar_condicional #'numberp '(A (1 B) ((2) NIL C) 3 (D (E 4)))) -> (1 2 3 4)
(defun aplanar_condicional (f_cond l)
    (cond
        ((null l)nil)
        ((listp (car l))
            (append(aplanar_condicional f_cond (car l))(aplanar_condicional f_cond (cdr l)))
        )
        ((funcall f_cond (car l))
            (list
                (cons (car l)(aplanar_condicional f_cond (cdr l)))
            )
        )
        (t
            (aplanar_condicional f_cond (cdr l))
        )
    )
)

;Sin usar recursividad y utilizando puramente lógica funcional, implementa la función fusion_columnas.
;Recibe una lista de funciones de longitud $N$, y un número arbitrario de listas de datos (todas también de longitud $N$). Debe devolver una única lista de longitud $N$, donde el elemento en la posición $i$ es el resultado de aplicar la función de la posición $i$ a todos los elementos en la posición $i$ de las listas de datos.
;(fusion_columnas '(#'+ #'* #'max) '(1 2 3) '(10 20 30) '(100 200 300)) -> (111 8000 300)
;; (+ 1 10 100) -> 111
;; (* 2 20 200) -> 8000
;; (max 3 30 300) -> 300
(defun fusion_columnas (lf &rest l)
    (apply #'mapcar (lambda (f &rest datos)(apply f datos)) lf l)
)

;Sin utilizar mapcar ni similares, implementa de forma puramente recursiva la función agrupar_iguales.
;Recibe una lista plana y debe devolver una lista de listas, donde los elementos consecutivos idénticos estén agrupados en la misma sublista.
;(agrupar_iguales '(A A B C C C A D D)) -> ((A A) (B) (C C C) (A) (D D))
(defun agrupar_iguales (l)
    (cond
        ((null l) nil)
        ((= (car l)(caar l))
            (cons
                (cons
                    (car l)(car(agrupar_iguales (cdr l)))(cdr(agrupar_iguales (cdr l)))
                )
            )
        )
        (t
            (cons
                (list(car l))(agrupar_iguales (cdr l))
            )
        )
    )
)

;Sin usar recursividad y utilizando puramente lógica funcional, implementa la función transponer.
;Recibe una única lista que representa una matriz (una lista que contiene sublistas, todas de la misma longitud). Debe devolver la matriz transpuesta (las filas se convierten en columnas)
;(transponer '((1 2 3) (4 5 6))) -> ((1 4) (2 5) (3 6))
(defun trasponer (m)
    (apply #'mapcar #'list m)
)

;Sin usar funciones de orden superior y usando exclusivamente recursividad pura, implementa la función podar_arbol.
;Recibe un valor objetivo y una lista de profundidad arbitraria. Debe devolver una copia exacta de la estructura original, pero eliminando completamente cualquier aparición del valor objetivo en cualquier nivel, sin dejar huecos vacíos (NILs no deseados).
;(podar_arbol 'X '(A (X B) ((X) NIL C) X (D (E X)))) -> (A (B) (NIL C) (D (E)))
(defun podar_arbol (n l)
    (cond 
        ((null l) nil)
        ((listp (car l))
            (cons(podar_arbol n (car l))(podar_arbol n (cdr l)))
        )
        ((= n (car l))
            (podar_arbol n (cdr l))
        )
        (t
            (cons(car l)(podar_arbol n (cdr l)))
        )
    )
)

;Sin utilizar mapcar y de forma puramente recursiva, implementa la función descomprimir.
;Recibe una lista plana compuesta por pares lógicos: un número entero seguido de un elemento. Debe devolver una lista con los elementos repetidos tantas veces como indique su número asociado. Si el número es 0, el elemento no se incluye.
(defun descomprimir (l)
    (cond
        ((null l) nil)
        ((> (car l) 0)
            (cons
                (cadr l)(descomprimir (cons (1- (car l)) (cdr l)))
            )
        )
        (t
            (descomprimir (cddr l)) 
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada funsmax que, dada una
;lista de valores y un número arbitrario de funciones, devuelva una lista cuyos elementos sean el máximo
;de aplicar cada una de las funciones sobre cada uno de los valores de la lista suministrada como
;parámetro
(defun funsmax (l &rest f)
    (mapcar (lambda(elemento)(apply #'max (mapcar (lambda(funcion)(funcall funcion elemento))f)))l)
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada generar_pares que, dadas
;dos funciones y un número arbitrario de argumentos, devuelva una lista de pares de elementos en
;formato de lista, donde cada par contenga el resultado de aplicar cada una de las dos funciones sobre el
;elemento correspondiente.
;(generarpares #’1+ (lambda(x)(* 2 x)) 1 2 3 4) -> ((2 2) (3 4) (4 6) (5 8))
;Implementa una versión recursiva que no utilice mapcar y otra basada en el uso de mapcar

;Version mapcar
(defun generar_pares (f g &rest a)
    (mapcar (lambda(x)(list (funcall f x)(funcall g x)))a)
)

;Version recursiva
(defun generar_pares (f g &rest a)
    (cond
        ((null a) nil)
        (t
            (cons
                (list(funcall f (car a))(funcall g (car a))(apply #'generar_pares f g (cdr a)))
            )
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada por_dos que, dada una
;lista que pudiera ser o no anidada, multiplique por 2 cada uno de sus elementos, manteniendo la
;estructura de la lista.
;(por_dos ‘(1 (2 3 (3 4)) 4 (2 3)))  (2 (4 6 (6 8)) 8 (4 6))
(defun por_dos (l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (por_dos (car l))(por_dos (cdr l)))
        )
        (t
            (cons(* 2 (car l)(por_dos (cdr l))))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada funmax que, dadas dos
;funciones y un número arbitrario de parámetros, devuelva una lista cuyos elementos sean el máximo
;entre aplicar la primera y la segunda función sobre cada uno de los argumentos.
;(funmax #’1+ (lambda(x) (* 2 x)) 0 1 2 3 4)  (1 2 4 6 8)
;(funmax (lambda(x) (+ 3 x)) (lambda(x) (* 2 x)) 0 1 2 3 4)  (3 4 5 6 8)
;Implementa una versión recursiva que no utilice mapcar y otra basada en el uso de mapcar

;version mapcar
(defun funmax (f g &rest p)
    (mapcar (lambda (parametro)(max (funcall f parametro)(funcall g parametro)))p)
)

;version recursiva
(defun funmax (f g &rest p)
    (cond
        ((null p) nil)
        (t
            (cons (max (funcall f (car p))(funcall g (car p)))
                (apply #'funmax f g (cdr p)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada suma que, dada una lista
;que pudiera ser o no anidada, calcule la suma de todos los elementos que contiene.
;(suma ‘(1 (2 3 (3 4)) 4 (2 3)))  22
;(suma ‘(1 (2 2) 3))  8
(defun suma (l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (+ (suma (car l))(suma (cdr l)))
        )
        (t
            (+ (car l)(suma (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplicar que, dada una
;lista sin anidamientos y un número arbitrario de funciones unarias, aplique secuencialmente las
;funciones sobre cada elemento de la lista.
;(aplicar ‘(1 2 3) #’1+)  (2 3 4)
;(aplicar ‘(1 2 3) #’1+ #’1+)  (3 4 5)
;(aplicar ‘(1 2 3) #’1+ #’1+ (lambda(x)(*2 x)))  (6 8 10)
(defun aplicar (l &rest f_u)
    (cond
        ((null f_u) l)
        (t
            (apply #'aplicar (mapcar (car f_u) l) (cdr f_u))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada maximo_anidado que, dada una lista que pudiera ser anidada, formada solo por números y sin listas vacías, devuelva el mayor de todos los valores en cualquier nivel de anidamiento. Proporciona dos versiones: una recursiva sin mapcar y otra basada en mapcar.
;version recursiva
(defun maximo_anidado (l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (max(maximo_anidad (car l)) (maximo_anidado (cdr l)))
        )
        (t
            (max (car l)(maximo_anidado (cdr l)))
        )
    )
)

;version mapcar
(defun maximo_anidado (l)
    (apply #'max (mapcar (lambda (x)(if (listp x)(maximo_anidado x) x ))l))
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada decrementar que, dada una lista que pudiera ser o no anidada, reste 1 a cada uno de sus elementos, manteniendo la estructura de la lista
(defun decrementar (l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (decrementar (car l)) (decrementar (cdr l)))
        )
        (t
            (cons(1- (car l))(decrementar (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada funmin que, dadas dos funciones y un número arbitrario de parámetros, devuelva una lista cuyos elementos sean el mínimo entre aplicar la primera y la segunda función sobre cada uno de los argumentos. Implementa una versión recursiva que no utilice mapcar y otra basada en el uso de mapcar.
;version recursiva
(defun funmin (f g &rest p)
    (cond
        ((null p) nil)
        (t
            (cons(min (funcall f (car p))(funcall g (car p)))
                (apply #'funmin f g (cdr p)))
        )
    )
)

;version mapcar
(defun funmin (f g &ret p)
    (mapcar (lambda (parametro)(min (funcall f parametro)(funcall g parametro))) p)
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada contar_atomos que, dada una lista que pudiera ser o no anidada, devuelva el número de átomos que contiene en cualquier nivel.
(defun contar_atomos (l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (+(contar_atomos (car l))(contar_atomos (cdr l)))
        )
        (t
            (1+ (contar_atomos (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplicarSobreUltimos que, dada una función y un número arbitrario de listas no vacías, devuelva una lista con el resultado de aplicar la función sobre el último elemento de cada lista.
(defun aplicarSobreUltimos (f &rest l)
    (cond
        ((null l) nil)
        (t
            (cons (funcall f (car (reverse car l)))
                (apply #'aplicarSobreUltimos f (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada maxMinProdLen que, dado un número arbitrario de argumentos numéricos (al menos uno), devuelva una lista de cuatro elementos con el mayor, el menor, el producto y el número de argumentos.
(defun maxMinProdLen (&rest a)
    (cond
        ((null a)(list nil nil 0 0))
        (t
            (list
                (apply #'max a)
                (apply #'min a)
                (apply #'* a)
                (length a)
            )
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada sumaCubos que, dada una lista sin anidamientos, devuelva la suma de los cubos de los elementos que contiene.
(defun sumaCubos (l)
    (cond
        ((null l) 0)
        (t
            (+ (* (car l) (car l) (car l))(sumaCubos (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada sumas que, dado un número arbitrario de listas no anidadas, devuelva la suma de cada una de ellas en forma de lista.
(defun sumas (&rest l)
    (cond
        ((null l) nil)
        (t
            (cons (apply #'+ (car l))(apply #'sumas (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplicar_en_impares que tome como entrada una función unaria y un número arbitrario de valores numéricos, y devuelva una lista con los argumentos en los que la función se ha aplicado solo sobre los que ocupan posiciones impares (1.º, 3.º…).
(defun aplicar_en_impares (f &rest a)
    (cond
        ((null a) nil)
        ((null (cdr a))
            (cons (funcall f (car a))(apply #'aplicar_en_impares f (cdr a)))
        )
        (t
            (cons (funcall (car a))(cons (cadr a)(apply #'aplicar_en_impares f (cddr a))))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplicar_pares que tome como entrada una función binaria y un número arbitrario de valores numéricos, y devuelva una lista con los resultados de aplicar la función sobre los argumentos dos a dos. Si el número de valores es impar, el último se deja tal cual.
(defun aplicar_pares (f &rest a)
    (cond
        ((null a) nil)
        ((null (cdr a))
            (cons (car a)(apply #'aplicar_pares f (cdr a)))
        )
        (t
            (cons (funcall f (car a) (cadr a))(apply #'aplicar_pares f (cddr a)))
        )
    )       
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada separar_paridad que, dado un número arbitrario de argumentos enteros, devuelva una lista de dos elementos: una lista con los pares y otra con los impares. Proporciona dos soluciones: la primera usando mapcar y sin recursión, y la segunda recursiva y sin mapcar.
;version recursiva
(defun separar_paridad (&rest a)
    (cond
        ((null a)(list nil nil))
        ((= (mod (car a) 2) 0)
            (list 
                (cons (car a) (car (apply #'separar_paridad (cdr a))) 
                    (cadr (apply #'separar_paridad (cdr a))))
            )
        )
        (t
            (list 
                (car (apply #'separar_paridad (cdr a)))
                    (cons (cadr (apply #'separar_paridad (cdr a))))
            )
        )
    )   
)

;version mapcar
(defun separar_paridad (&rest a)
    (list
        (remove nil (mapcar (lambda (x)(if(= 0 (mod x 2)) x nil))a))
        (remove nil (mapcar (lambda (x)(if(= 0 (mod x 2)) nil x))a))
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada procesar_rotando que, dada una lista de valores y un número arbitrario de funciones, devuelva una lista donde las funciones se hayan ido aplicando de forma rotativa sobre cada elemento: la primera sobre el primer elemento, la segunda sobre el segundo, y al agotarlas se vuelve a la primera. Si no hay funciones se devuelve la lista original.
(defun procesar_rotando (l &rest f)
    (cond
        ((null l) nil)
        ((null f) l)
        (t
            (cons (funcall (car f) (car l))(apply #'procesar_rotando (cdr l) (append (cdr f) (list (car f)))))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada generar_triplas que, dadas tres funciones y un número arbitrario de argumentos, devuelva una lista de listas de tres elementos, donde cada una contenga el resultado de aplicar cada una de las tres funciones sobre el argumento correspondiente. Implementa una versión recursiva que no utilice mapcar y otra basada en mapcar.
;version recursiva
(defun generar_triplas (f g h &rest a)
    (cond
        ((null l) nil)
        (t
            (cons
                (list
                    (funcall f (car a))(funcar g (car a))(funcall h (car a))
                        (apply #'generar_triplas f g h (cdr a))
                )
            )
        )
    )
)

;version mapcar
(defun generar_triplas (f g h &rest a)
    (list
        (mapcar (lambda (x)(list (funcall f x)(funcall g x)(funcall h x)))a)
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada segundo_y_penultimo que, dada una lista sin anidamientos, devuelva una lista con su segundo y su penúltimo elemento. Si alguno no existe, será NIL.
(defun segundo_y_penultimo (l)
    (list (cadr l) (cadr (reverse l)))
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada suma_anidada que, dada una lista que pudiera ser o no anidada y que contiene solo números, devuelva la suma de todos ellos, en cualquier nivel de anidamiento.
(defun suma_anidada (l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (+ (suma_anidada (car l)) (suma_anidada (cdr l)))
        )
        (t
            (+ (car l)(suma_anidada (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada maxmin_f que, dadas dos funciones y un número arbitrario de argumentos, devuelva una lista de listas de dos elementos. Para cada argumento, la lista contendrá el mayor y el menor de los resultados de aplicar la primera y la segunda función sobre él. Implementa una versión recursiva que no utilice mapcar y otra basada en el uso de mapcar.
;version recursiva
(defun maxmin_f (f g &rest a)
    (cond
        ((null a) nil)
        (t
            (cons
                (list(max (funcall f (car a))(funcall g (car a)))(min (funcall f (car a))(funcall g (car a))))
                        (apply #'maxmin_f f g (cdr a))
            )
        )
    )
)

;version mapcar
(defun maxmin_f (f g &rest a)
    (mapcar (lambda (x)(list (max (funcall f x)(funcall g x))
                                (min (funcall f x)(funcall g x))))a)
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada minimos que, dado un número arbitrario de listas no anidadas del mismo tamaño, devuelva una lista con el menor valor de cada posición, es decir, el mínimo de los primeros elementos, el de los segundos, etc.
(defun minimos (&rest l)
    (apply #'mapcar #'min l)
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplicar_anidada que, dada una función unaria y una lista que pudiera ser o no anidada y que contiene solo números, devuelva una lista con la misma estructura en la que la función se haya aplicado a cada uno de los números.
(defun aplicar_anidada (f_u l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (aplicar_anidada f_u (car l))(aplicar_anidada f_u (cdr l)))
        )
        (t
            (cons (funcall f_u (car l))(aplicar_anidada f_u (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales y usando mapcar, implementar una función en lisp llamada contar_cumplen que, dada una lista no anidada de números y un número arbitrario de funciones que representan condiciones, devuelva una lista con cuántos elementos de la lista cumplen cada una de las condiciones, en el mismo orden que las funciones.
(defun contar_cumplen (l &rest f)
    (mapcar (lambda (funcion)
        (length(remove nil(mapcar (lambda (x)(if (funcall funcion x)x nil))l))))f)
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada suma_extremos que, dada una lista de listas no anidadas y no vacías, devuelva una lista con la suma del primer y el último elemento de cada una de ellas.
(defun suma_extremos (l)
    (cond
        ((null l) nil)
        (t
            (cons(+(caar l)(car(reverse (car l))))(suma_extremos (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada contar_elemento que, dado un elemento y una lista que pudiera ser o no anidada, devuelva cuántas veces aparece ese elemento en la lista, en cualquier nivel de anidamiento.
(defun contar_elemento (e l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (+(contar_elemento e (car l))(contar_elemento e (cdr l)))
        )
        ((equal e (car l))
            (+ 1 (contar_elemento e (cdr l)))
        )
        (t
            (contar_elemento e (cdr l))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada suma_productos que, dadas dos funciones y un número arbitrario de argumentos, devuelva la suma, para todos los argumentos, del producto de aplicar la primera función por el resultado de aplicar la segunda. Proporciona dos soluciones: una recursiva sin mapcar y otra basada en mapcar.
;version recursiva
(defun suma_productos (f g &rest a)
    (cond
        ((null a) 0)
        (t
            (+(*(funcall f (car a))(funcall g (car a)))(apply #'suma_productos f g (cdr a)))
        )
    )
)

;version mapcar
(defun sum_prodcutos (f g &rest a)
    (apply #'+ (mapcar (lambda (x)(* (funcall f x)(funcall g x)))a))
)

;Sin definir otras funciones adicionales y usando mapcar, implementar una función en lisp llamada distancia2 que, dadas dos listas no anidadas de números del mismo tamaño, devuelva la suma de los cuadrados de las diferencias entre los elementos que ocupan la misma posición.
;version recursiva
(defun distancia2 (l1 l2)
    (cond
        ((null l1) 0)
        (t
            (+(*(- (car l1) (car l2))(- (car l1) (car l2)))(distancia2 (cdr l1) (cdr l2)))
        )
    )
)

;version mapcar
(defun distancia2 (l1 l2)
  (apply #'+ (mapcar (lambda (a b) (* (- a b) (- a b))) l1 l2))
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada quitar que, dado un elemento y una lista que pudiera ser o no anidada, devuelva la lista sin ninguna aparición de ese elemento, en cualquier nivel de anidamiento y manteniendo la estructura del resto. Si una sublista se queda vacía, se conserva como NIL.
(defun quitar (e l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (quitar e (car l))(quitar e (cdr l)))
        )
        ((equal e (car l))
            (quitar e (cdr l))
        )
        (t
            (cons (car l)(quitar e (cdr l)))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada mayor_resultado que, dada una función unaria y un número arbitrario (al menos uno) de argumentos, devuelva el mayor de los resultados de aplicar la función a cada uno de los argumentos. Proporciona dos soluciones: una recursiva sin mapcar y otra basada en mapcar.
;version recursiva
(defun mayor_resultado (f_u &rest a)
    (cond
        ((null (cdr a)) (funcall f_u (car a)))
        (t
            (max (funcall f_u (car a))(apply #'mayor_resultado f_u (cdr a)))
        )
    )
)

;version mapcar
(defun mayor_resultado (f_u &rest a)
    (apply #'max (mapcar (lambda (x)(funcall f_u x))a))
)

;Sin definir otras funciones adicionales y usando mapcar, implementar una función en lisp llamada longitud_maxima que, dado un número arbitrario de listas no anidadas, devuelva la longitud de la más larga.
(defun longitud_maxima (&rest l)
    (apply #'max (mapcar #'length l))
)

;Sin definir otras funciones adicionales, escribe una función en lisp llamada sumafunciones que acepte un número indeterminado de funciones unarias seguidas de un valor numérico y devuelva la suma resultante de la aplicación de cada una de las funciones sobre el valor.
(defun sumafunciones (v &rest f)
    (apply #'+ (mapcar (lambda (funcion)(funcall funcion v))f))
)

;Sin definir otras funciones auxiliares, define una función en lisp llamada elemento_medio que, dada una función y una lista de enteros, devuelva como resultado la aplicación de la función sobre su elemento central, o NIL si el número de elementos fuera par.
(defun elemento_medio (f l)
    (cond
        ((null l) nil)
        ((null (cdr l))(funcall f (car l)))
        (t
            (elemento_medio f (cdr (reverse (cdr (reverse l)))))
        )
    )
)

;Sin utilizar funciones adicionales, implementa una función en lisp llamada filtrar que, dada una lista, una posición y una función (en ese orden, como en los ejemplos), devuelva una nueva lista en la que:
;se haya eliminado el elemento que ocupe la posición especificada (la primera posición es la 1), y
;se haya añadido como último elemento el resultado de aplicar la función sobre el elemento eliminado.
(defun filtrar (l p f)
    ((null l) nil)
    ((= p 1 )
        (append (cdr l)(list funcall (car l)))
    )
    (t
        (cons (car l)(filtrar (cdr l) (- p 1) f))
    )
)

;Sin definir otras funciones adicionales, implementa una función en lisp llamada pasarABinario que, dado un número en ddecimal, devuelva los valores que componen su representación binaria en forma de lista.
(defun pasarABinario (n)
    (cond
        ((< n 2) (list n))
        (t
            (append (pasarABinario (/ (- n (mod n 2)) 2)) (list (mod n 2)))
        )
    )
)

;Sin definir otras funciones auxiliares, escribe una función en lisp llamada empareja que, dada una lista sin anidamientos, devuelva una nueva lista cuyos elementos sean pares que contengan el primer y último elemento, el segundo y el penúltimo, y así sucesivamente. Si la lista contiene un número impar de elementos, el último par consistirá del elemento central replicado.
(defun empareja (l)
    (cond
        ((null l) nil)
        ((null cdr)(list(list (car l)(car l))))
        (t
            (cons
                (list(car l) (car(reverse l))(emparejar (reverse (cdr (reverse (cdr l))))))
            )
        )
    )
)