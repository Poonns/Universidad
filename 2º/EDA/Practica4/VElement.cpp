/**
* Practicas de EDA.
 * Curso 2012-2013
 * Departament d'Informatica.
 * Universitat de Valencia.
 */

#include "VElement.h"


/**
 * Clase de elementos de vector que mantiene contador estático de intercambios.
 *
 */
int VElement::nAssign = 0;
int VElement::nComparison = 0;

VElement::VElement()
{
    value = 0;
}

VElement::VElement(int v)
{
    value = v;
    nAssign++;
}

VElement::VElement(const VElement &vE)
{
    value = vE.value;
    nAssign++;
}

void VElement::set(int v)
{
    value = v;
}

int VElement::get() const
{
    return value;
}

bool VElement::operator >(const VElement &vE) const
{
    nComparison++;
    return (get() > vE.get());
}

bool VElement::operator >(int e) const
{
    nComparison++;
    return (get() > e);
}

bool VElement::operator <(const VElement &vE) const
{
    nComparison++;
    return (get() < vE.get());
}

bool VElement::operator <(int e) const
{
    nComparison++;
    return (get() < e);
}

bool VElement::operator >=(const VElement &vE) const
{
    nComparison++;
    return (get() >= vE.get());
}
bool VElement::operator >=(int e) const
{
    nComparison++;
    return (get() >= e);
}

bool VElement::operator <=(const VElement &vE) const
{
    nComparison++;
    return (get() <= vE.get());
}

bool VElement::operator <=(int e) const
{
    nComparison++;
    return (get() <= e);
}

bool VElement::operator ==(const VElement &vE) const
{
    nComparison++;
    return (get() == vE.get());
}

bool VElement::operator ==(int e) const
{
    nComparison++;
    return (get() == e);
}

VElement & VElement::operator =(int e)
{
    value = e;
    nAssign++;
    return *this;
}

VElement & VElement::operator =(const VElement &vE)
{
    value = vE.value;
    nAssign++;
    return *this;
}


