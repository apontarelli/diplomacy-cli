5. THE PROCESS OF ADJUDICATION
The article "The Math of Adjudication" of the Diplomatic Pouch, Spring 2009 Movement contains additional information with more examples. However, this chapter was rewritten for version 3.0 of the DATC and is more recent than the article.

Writing a Diplomacy adjudicator program may look not more difficult than writing a program that checks the moves of a chess game. However, the contrary is true. A Diplomacy adjudicator that passes all test cases as described in this document contains many small and difficult details.

The first thing people think about when analyzing the rules of Diplomacy is the sequence in which the orders are processed. However, this is exactly that should not be done. None of the rulebooks describe or hint about the sequence in which the orders must be adjudicated. The rules are a set of "equations" or "conditions" that should be fulfilled. The first step is to make these equations more precise, more mathematically.

A key feature of these equations is that the result during adjudication will not change. An order succeeds or fails and that does not change during adjudication. However, when the adjudication is started its state is "unknown" or "unresolved". This in contrast to principles like a support is successful until it is cut or a move order succeeds until it bounces. Going that way makes writing a correct adjudicator very difficult.

When the equations are understood, the next step is to make an algorithm that finds a solution.

5.A. OVERVIEW OF DIPLOMACY EQUATIONS
There are three active orders that should be adjudicated in success or fail. The result can be kept in a Boolean, were success is true and fail is false.

MOVE
SUPPORT
CONVOY
A unit is dislodged when it doesn't move and another unit successfully moves to the area. This doesn't need further explanation.

In further detailing out the conditions, it is possible to refer to the result of another order, even when the result of that order is not known yet. The sequence of adjudication of orders will be discussed later.

There are a number of different strengths that need to be calculated during adjudication. They may result in the same value, however may differ in some circumstances. They should not be confused.

ATTACK STRENGTH
For each unit ordered to move, the strength to attack and conquer the area moved to.
HOLD STRENGTH
For each area on the board the strength to prevent that a unit moves to that area. If the area is empty, the value is zero.
DEFEND STRENGTH
The strength of unit ordered to move engaged in a head-to-head to prevent that the opposing unit succeeds.
PREVENT STRENGTH
The strength of unit ordered to move preventing another unit ordered to succeed moving to the same area.
Typically, if a move is not engaged in a head-to-head battle the ATTACK STRENGTH must be greater than the HOLD STRENGTH of the area and the PREVENT STRENGTH of each of the units competing for the same area, to be successful.

If multiple units are ordered to move to the same area, then each order has its own success condition. So, there is not a mechanism to declare a grand winner of the contested area.

Example:
Germany:
A Berlin -> Silesia
A Munich Supports Berlin -> Silesia

Russia:
A Warsaw -> Silesia

Austria:
A Bohemia -> Silesia
The success of the MOVE orders of Berlin, Warsaw and Bohemia are determined separately. The orders of Warsaw and Bohemia fail, because their ATTACK STRENGTH is one, while the PREVENT STRENGTH of Berlin is two. The MOVE order of Berlin succeeds, because the ATTACK STRENGTH is two and that beats the PREVENT STRENGTH of both the units in Warsaw and Bohemia.

Similarly, in case of a head-to-head battle the MOVE of both units is determined separately.


Example:
Germany: 
A Berlin - Kiel
A Munich Supports F Kiel - Berlin

England: 
F Kiel - Berlin
In this example Germany probably reversed the support order in Munich by mistake. The MOVE of Berlin will fail, because the ATTACK STRENGTH is one, while the DEFEND STRENGTH of Kiel is two. The MOVE of Kiel will also fail, because the ATTACK STRENGTH of one is insufficient to beat the DEFEND STRENGTH of Berlin also on. Since Munich may not help in dislodging a unit of the same nationality, but may help in defending, the ATTACJ STRENGTH and DEFEND STRENGTH of Kiel are differently.

Finally, we have the PATH Boolean variable. For a convoying army it determines whether there is an uninterrupted path to the destination.

5.B. PRECISE DESCRIPTION OF THE EQUATIONS
With the general idea of the different terms given in the previous section, the conditions can be further detailed out. Note, that various rule issues, such as coasts or pre-scanning the orders are not covered here. They are not relevant for the algorithm.

5.B.1. MOVE
In case of a head-to-head battle, the move succeeds when the ATTACK STRENGTH is larger than the DEFEND STRENGTH of the opposing unit and larger than the PREVENT STRENGTH of any unit moving to the same area. If one of the opposing strengths is equal or greater, then the move fails.
If there is no head-to-head battle, the move succeeds when the ATTACK STRENGTH is larger than the HOLD STRENGTH of the destination and larger than the PREVENT STRENGTH of any unit moving to the same area. If one of the opposing strengths is equal or greater, then the move fails.
5.B.2. SUPPORT
A support order is cut (fails) when another unit is ordered to move to the area of the supporting unit and the following conditions are satisfied:

The moving unit has a successful PATH.
The moving unit is of a different nationality.
The destination of the supported unit is not the area of the unit attacking the support.
Or:

The unit successfully moves (dislodging the supporting unit).
5.B.3. CONVOY
The convoy order is successful, if the fleet is not dislodged. So, no other unit moves successfully to the area.

5.B.4. PATH
The PATH of a move order is successful when the unit can directly move to the destination and is not convoyed or is convoyed and there is a chain of adjacent fleets from origin to destination each with a matching and successful CONVOY order.

5.B.5. HOLD STRENGTH
The HOLD STRENGTH is defined for an area, rather than for an order.
The hold strength is zero when the area is empty, or when it contains a unit with successful MOVE order.
It is one when the area contains a unit with a failed MOVE ORDER
In all other cases, it is one plus the number of units support it in hold with successful SUPPORT order.
5.B.6. PREVENT STRENGTH
If the PATH of the move order fails, then the PREVENT STRENGTH is 0.
In cases where the move is part of a head-to-head battle and the MOVE order of the opposing unit is successful, then the prevent strength is 0.
In the remaining cases the PREVENT STRENGTH is 1 plus the number of units with a successful SUPPORT order.
5.B.7. DEFEND STRENGTH
The DEFEND STRENGTH of a unit with a move order is one plus the number of units supporting the move with a successful SUPPORT order.

5.B.8. ATTACK STRENGTH
If the PATH of the move order fails, then the ATTACK STRENGTH is zero.
Otherwise, if the destination is empty, or in a case where there is no head-to-head battle and the unit at the destination has a move order for which the move is successful, then the ATTACK STRENGTH is one plus the number of units supporting the move with successful SUPPORT order.
If not and the unit at the destination is of the same nationality, then the ATTACK STRENGTH is zero.
In all other cases, the ATTACK STRENGTH is one plus the number of units supporting the move, of different nationality as the unit on the destination and with successful SUPPORT order.
5.B.9. CIRCULAR MOVEMENT AND PARADOXES
It is possible that for an order set there is no solution for given condition or that there are multiple solutions. This is only possible when there is a circular dependency between the different conditions. That happens with circular movement or with a convoy paradox.

However, a circular dependency may also have just one solution. For instance, in case of a circular movement, if one the unit moves for sure because of a support or a unit won't move because of a bounce, the circular movement has only one solution. Similar, a convoy paradox might lose its paradox status if an additional support or bounce is added. If there is only one solution that solution must be taken.

In case of a circular dependency that has zero or two solutions the orders that are in the circular dependency must be examined. If they only consist of move orders, then it is a circular movement. All moves become successful. If there are any convoying fleets, then there is a convoy paradox. The Szykman rule must be applied and the CONVOY orders of the convoying fleets fail (ignoring the original condition for CONVOY order).

In case the adjudicator can also handle variants, other paradoxes may apply that do not fulfil above conditions. In such case it is best to let all orders in the cycle fail.

5.C. FROM CONDITIONS TO ALGORITHM
The next step is to transform the conditions as described in the previous sections to an algorithm. When finished, the algorithm can be tested with the test cases. Note, that if a bug is found, then it is unlikely to be a problem of the conditions described in the previous section. They have proven to be stable from the first version.

The adjudication program needs to handle the following situations:

An order that is not indirectly dependent on itself.
An order that is indirectly dependent on itself, but there is still exactly one resolution.
An order that is indirectly dependent on itself and there are zero or two resolutions.
If it was only needed to adjudicate orders of category a, the algorithm would be simple. Just a recursive function that adjudicates an order and if it depends on a different order, it will call itself recursively. This also shows that it is not needed to be concerned about the sequence of orders. Rule of thumbs such as cut support first are not needed to be programmed.

It is also not difficult to detect whether the recursive hits a cyclic dependency. Applying the backup rule to the cycle will also not give too much trouble. The real challenge is to determine whether the cycle has just one solution as in category b (that solution should then be taken) or zero or two solutions as in category c (the backup rule should be applied).

There are two approaches to this problem:

Making decision on partial information.
Guessing different resolutions.
Both can lead to a correct algorithm. The difference can be best explained by an example:

Russia:
A Constantinople - Smyrna

Turkey: 
F Ankara - Constantinople
A Smyrna - Ankara
A Bulgaria Supports F Ankara - Constantinople
This circular movement has only solution, because MOVE of Ankara to Constantinople is guaranteed to succeed, due to the support. And this is exactly how an algorithm based on partial information works. Even if the MOVE order of Constantinople to Smyrna is still uncertain, it can be concluded that the HOLD STRENGTH of Constantinople is zero or one, but not higher. The ATTACK STRENGTH of Ankara to Constantinople is two and will always beat this allowing the MOVE to succeed. When this is concluded, the other MOVE orders can also succeed.

The guessing algorithm just makes a guess for one of the orders. For instance, it guesses that the MOVE from Smyrna to Ankara fails. As result the MOVE from Constantinople to Smyrna fails, but the MOVE from Ankara to Constantinople still succeeds. If then the MOVE of the Smyrna is adjudicated it is concluded that it succeeds which is inconsistent with the initial guess. The guess can be repeated with a successful MOVE and then a consistent adjudication is obtained.

For both approaches it is suggested to make two mutual recursive functions (in Python code):

def adjudicate(order):
    ...

def resolve(order):
    ...
Both functions take an order reference as input (in other languages this can be an index or pointer) and return as Boolean the success or failure of the order.

The "adjudicate" function implements the conditions as discussed earlier. The function can be split up in adjudicate functions for MOVE, SUPPORT and CONVOY and separate functions for PATH and STRENGTH values. If the result of another order is required, then it will call the "resolve" function. The "adjudicate" function will not update any administration.

The "resolve" function is a generic function and has no knowledge about the details of the orders. It updates the administration when an adjudication becomes final, prevents that the same order is adjudicated twice, detect cyclic dependencies and applies the backup rule if needed.

The adjudicate function will be roughly the same for both approaches, while the resolve function will be fundamentally different.

The final program is then just calling resolve for each order to ensure that every order is resolved.

5.D. THE TROUBLE WITH PANDIN'S PARADOX
For solving circular dependencies, it would be great if each order is part of at most one cycle. However, in convoy paradoxes, it can be complex. Consider Pandin's paradox:

England: 
F London Supports F Wales - English Channel
F Wales - English Channel

France: 
A Brest - London
F English Channel Convoys A Brest - London

Germany: 
F North Sea Supports F Belgium - English Channel
F Belgium - English Channel
The cycle of dependencies looks like this:

English Channel
Wales
Belgium
London
The arrows are in the direction of going deeper in recursion. To decide whether the convoying fleet in the English succeeds it is needed to know whether the fleet in Wales or Belgium succeeds. For calculating the success of Wales and Belgium, the result of the SUPPORT of London is required. And the SUPPORT of London depends on the successful CONVOY of the English Channel. Note, the MOVE of the army in Brest is a key ingredient of this paradox, but is not listed in the cycle. The success of this order is not relevant (it always fails), the question is whether it will cut the SUPPORT of London.

The fact that there is not a single cycle, makes the algorithms complicated. In case the algorithm can make decisions on partial information, then it can conclude that the unit in Wales will never succeed, because it will never have enough ATTACK STRENGTH against the unit in Belgium. However, the algorithm should ensure that all decisions on partial information are finished, before acting on cycles.

For a guessing algorithm these situations are very difficult to handle. For instance, it should not start making a guess on Wales, because that order is not a key decision that influences the paradox.

There is a simple hack that ensures that only clean single cycle can happen. The adjudicate function for CONVOY would normally call resolve for any units attacking the convoy. That make the success of the fleet in the English Channel dependent on Wales and Belgium in above example. However, if the adjudicate function is called instead, the dependency is skipped and all cycles will be clean and simple. In the above example the cycle will only consists out of English Channel and London. If the Szykman rule is applied on that cycle, the English Channel will fail to convoy and all other orders can be resolved.

5.E. THE PARTIAL INFORMATION ALGORITHM
If the success or failure of any dependent order is still uncertain it might be needed to adjudicate an order twice. This can best be illustrated by an example:


England:
F North Sea - Holland
F Belgium Support F North Sea - Holland

Germany:
A Holland Hold
Suppose furthermore that the SUPPORT of Belgium cannot be determined yet. To see whether the North Sea MOVE succeeds, the ATTACK STRENGTH is calculated. This may not exceed 1 and this insufficient to dislodge Holland. However, this absence of success does not imply that the order fails. This can separately be checked.

In earlier descriptions of this algorithm the failure and success of an order in the adjudicate function was implemented separately, doubling the code. However, there is a neat trick to avoid that. Both adjudicate and resolve functions get an additional parameter:

def adjudicate(order, optimistic):
    ...

def resolve(order, optimistic):
    ...
If the optimistic parameter is True, then any uncertain information is assumed to have a value that is supportive for letting the order succeed. In the pessimistic scenario (the parameter is False) the opposite. STRENGTH values will return what the strength will be at most in the optimistic scenario and the least in the pessimistic scenario.

The adjudicate function will not check on the optimistic parameter, but just pass it to its subfunctions and to the resolve function, but sometimes inverse it. For instance, for a MOVE adjudication the ATTACK STRENGTH must be calculated. In the optimistic scenario the optimistic ATTACK STRENGTH is calculated but for the STRENGTH values of any opposing unit the pessimistic value is used. Similar, in the optimistic scenario of a SUPPORT or CONVOY order the pessimistic result is used for any unit that can let the order fail.

The resolve function (with cycle detection, but without the backup rule) becomes something like this:


def resolve(order, optimistic):
    if order.resolved:
        return order.resolution

    if order.visited:
        # We hit cyclic dependency.
        return optimistic # Success when optimistic,
                          # fail when pessimistic.

    order.visited = True # Prevent endless recursion.
    opt_result = adjudicate(order, True)
    pes_result = adjudicate(order, False)
    order.visited = False
    
    if opt_result == pes_result:
        # We have a single resolution.
        # Store the result and return it.
        order.resolution = opt_result
        order.resolved = True
        return opt_result

    # Order still undecided. Success when optimistic.
    return optimistic.
As you can see the resolve function does not just pass the optimistic parameter to the adjudicate function. It will try both optimistic and pessimistic scenarios and if they are in agreement the order is resolved.

In case there is only one resolution of the situation, then the resolve function will always resolve the order (although, the resolution of any depending orders may still stay open). To add the backup rule for cyclic movement and the Szykman rule, we need to track which orders are in the cycle when no resolution could be found. We do this by keeping those orders in a global array 'cycle'. Example code:

def resolve(order, optimistic):
    global cycle
    if order.resolved:
        return order.resolution

    if order in cycle:
        # We already concluded that this order is in a cycle
        # which we cannot yet resolve.
        return optimistic # Success when optimistic,
                          # fail when pessimistic.

    if order.visited:
        # We hit cyclic dependency.
        cycle.add(order)
        return optimistic

    order.visited = True # Prevent endless recursion.
    old_cycle_len = len(cycle)
    opt_result = adjudicate(order, True)
    pes_result = adjudicate(order, False)
    order.visited = False
    
    if opt_result == pes_result:
        # We have a single resolution.
        # Wipe out any cycle information that was found in
        # recursion.
        del cycle[old_cycle_len:]
        # Store the result and return it.
        order.resolution = opt_result
        order.resolved = True
        return opt_result

    if order in cycle:
        # We returned from recursion, where this order hit the
        # cycle and we didn't get a single resolution.
        # Apply backup rule on all those orders.
        backup_rule(cycle[old_cycle_len:])
        del cycle[old_cycle_len:]
        # The backup rule might not have resolved this order.
        return resolve(order, optimistic)

    # We are returning from a situation where a cycle was
    # detected. However, this order is not the ancestor of the
    # whole cycle. We further retreat from recursion.
    cycle.add(order)
    return optimistic
If the cycle is not a clean single cycle, it may fail in general, although not with the standard rules. If the resolve function is called on Wales as in the situation of 5.D, it will call in recursion, London, English Channel, Wales, Belgium and London before retreating from recursion. Then it will conclude that there is a convoy paradox when it is back by the London order. But at that moment, the Wales order is still undecided, while it could be decided by partial information. With the standard rules this still goes right, however, in case of variant rules things can be different. A way to fix this is to retreat in recursion up to the order that is the ancestor of the whole cycle (Wales in the example). This can be achieved by adding a global integer 'recursion_hits' that keeps track how many times the recursion hit previous visited orders:

def resolve(order, optimistic):
    global cycle, recursion_hits
    if order.resolved:
        return order.resolution

    if order in cycle:
        # We already concluded that this order is in a cycle
        # which we cannot yet resolve.
        return optimistic # Success when optimistic,
                          # fail when pessimistic.

    if order.visited:
        # We hit cyclic dependency.
        cycle.add(order)
        recursion_hits += 1
        return optimistic

    order.visited = True # Prevent endless recursion.
    old_cycle_len = len(cycle)
    old_recursion_hits = recursion_hits
    opt_result = adjudicate(order, True)
    pes_result = adjudicate(order, False)
    order.visited = False
    
    if opt_result == pes_result:
        # We have a single resolution. Wipe out any
        # cycle information that was found in recursion.
        del cycle[old_cycle_len:]
        recursion_hits = old_recursion_hits
        # Store the result and return it.
        order.resolution = opt_result
        order.resolved = True
        return opt_result

    if order in cycle:
        # We returned from recursion, where this order hit the
        # cycle and we didn't get a single resolution.
        recursion_hits -= 1

    if recursion_hits == old_recursion_hits:
        # We have sufficiently retreated from recursion such
        # that this order was the ancestor of the whole cycle.
        # Apply backup rule on all orders in cycle.
        backup_rule(cycle[old_cycle_len:])
        del cycle[old_cycle_len:]
        # The backup rule might not have resolved this order.
        return resolve(order, optimistic)

    # We are returning from a situation where a cycle was
    # detected. However, this order is not the ancestor of the
    # whole cycle. We further retreat from recursion.
    if not order in cycle:
        cycle.add(order)
    return optimistic
This is a very robust way of handling, especially in variants, since all orders that can be decided on partial or full information will decided that way (as players expect).

The disadvantage of this algorithm is that all orders are adjudicated twice, in optimistic and pessimistic mode. With modern powerful computers, this shouldn't be a problem, except when it is used in an AI engine. A simple optimization is to skip the pessimistic adjudication when the optimistic adjudication fails. In such case the pessimistic adjudication is guaranteed to fail also. A further optimization is to track whether the adjudication of an order was fully dependent on resolved orders. In such case, one adjudication suffice. This can be implemented by adding another global variable 'uncertain'. This variable is set to True if a dependent order does not have a resolution yet and kept unaltered otherwise.

def resolve(order, optimistic):
    global cycle, recursion_hits, uncertain
    if order.resolved:
        return order.resolution

    if order in cycle:
        # We already concluded that this order is in a cycle
        # which we cannot yet resolve.
        # Result is based on uncertain information.
        uncertain = True
        return optimistic # Success when optimistic,
                          # fail when pessimistic.

    if order.visited:
        # We hit cyclic dependency.
        cycle.add(order)
        recursion_hits += 1
        uncertain = True
        return optimistic

    order.visited = True # Prevent endless recursion.
    old_cycle_len = len(cycle)
    old_recursion_hits = recursion_hits
    old_uncertain = uncertain
    uncertain = False
    opt_result = adjudicate(order, True)
    # Try to avoid a second adjudication for performance.
    pes_result = adjudicate(order, False) if uncertain and opt_result else opt_result
    order.visited = False
    
    if opt_result == pes_result:
        # We have a single resolution. Wipe out any
        # cycle information that was found in recursion.
        del cycle[old_cycle_len:]
        recursion_hits = old_recursion_hits
        # The uncertain variable must be unaltered, because
        # order is resolved now.
        uncertain = old_uncertain
        # Store the result and return it.
        order.resolution = opt_result
        order.resolved = True
        return opt_result

    if order in cycle:
        # We returned from recursion, where this order hit the
        # cycle and we didn't get a single resolution.
        recursion_hits -= 1

    if recursion_hits == old_recursion_hits:
        # We have sufficiently retreated from recursion such
        # that this order was the ancestor of the whole cycle.
        # Apply backup rule on all orders in cycle.
        backup_rule(cycle[old_cycle_len:])
        del cycle[old_cycle_len:]
        uncertain = old_uncertain
        # The backup rule might not have resolved this order.
        return resolve(order, optimistic)

    # We are returning from a situation where a cycle was
    # detected. However, this order is not the ancestor of the
    # whole cycle. We further retreat from recursion.
    if not order in cycle:
        cycle.add(order)
    return optimistic
One should remind that one adjudicate call, may result in multiple resolve calls for the same order. For instance, for MOVE order the ATTACK STRENGTH is calculated and the HOLD STRENGTH of the destination. Both STRENGTH values depend on the MOVE on the unit on the destination. Especially in a cyclic movement, where orders cannot directly get a permanent resolution, there is a risk that one gets an exponential explosion. If the cyclic movement consists of 10 moves, they may lead to 2 to the power of 10 (1024) adjudications. In above resolve function this is prevented by the check 'if order in cycle:'. One could also program the adjudicate function in such way that double calls to resolve are not made.

5.F. THE GUESS ALGORITHM
In the guess algorithm different resolutions are tried and checked on consistency. The difficulty is to do this in a simple recursive function. The trick is to set the guess value when an order is visited for the first time and use this guess value when this order is visited again deeper in the recursion. If this happens orders are again added to the global variable 'cycle' when returning from the recursion.

def resolve(order):
    global cycle
    if order.resolved or order in cycle:
        # In case the order is in cycle, then the resolution
        # given earlier was stored in the resolution,
        # without setting the order to resolved.
        return order.resolution

    if order.visited:
        # We hit a cyclic dependency.
        cycle.add(order)
        return order.resolution

    order.visited = True # Prevent indefinite recursion.
    old_cycle_len = len(cycle)
    # We set the resolution to the value that is returned when
    # the same order is visited deeper in recursion.
    order.resolution = False
    first_result = adjudicate(order)

    if len(cycle) == old_cycle_len:
        # No cyclic dependencies were detected.
        # We can just take this resolution.
        order.resolution = first_result
        order.resolved = True
        return first_result

    if order in cycle:
        # Deeper in the recursion we hit a cycle on this order.
        # Try to adjudicate the cycle with different guess.    
        del cycle[old_cycle_len:]
        order.resolution = True # Was False on first run.
        second_result = adjudicate(order)

        if first_result == second_result:
            # Although we hit a cycle, there is only one result.
            del cycle[old_cycle_len:]
            order.resolution = first_result
            order.resolved = True
            return first_result

        # Different results on different guesses.
        backup_rule(cycle[old_cycle_len:])
        del cycle[old_cycle_len:]
        order.visited = False
        # Backup rule might not have resolved this order.
        return resolve(order)

    # We are returning from recursion where we hit a cycle,
    # but was not started on this order.
    cycle.add(order)
    # We remember the result if resolve is called again on same
    # order, while cycle is not yet resolved.
    order.resolution = first_result
    order.visited = False
    return first_result
Again, this algorithm is problematic when there is a cycle that is not a clean single cycle. This can be fixed with the hack described in section 5.D. There is no straight forward way to fix this in the generic resolve function in theoretical correct way. However, if the algorithm retreats to the situation where the order is the ancestor of the whole cycle, then in practice it will work.

If the resolve function is called on Wales as in the situation of 5.D, then the result will be negative, whether it guesses positive or negative, although the remaining paradox is not tried in two ways. If the resolve function is called on Belgium, London or English Channel, then the paradox is detected and Wales is added (incorrectly) to the cycle. However, the backup rule won't touch this order, so, it still goes right. For implementation, again a global variable 'recursion_hits' is added. But also a Boolean variable 'guess_based' is necessary, because looking whether the cycle has increased is not correct if multiple cycles are possible.

def resolve(order):
    global cycle, guess_based, recursion_hits
    if order.resolved
        return order.resolution

    if order in cycle:
        # In case the order is in cycle, then the resolution
        # given earlier was stored in the resolution,
        # without setting the order to resolved.
        guess_based = True
        return order.resolution

    if order.visited:
        # We hit a cyclic dependency.
        cycle.add(order)
        guess_based = True
        recursion_hits += 1
        return order.resolution

    order.visited = True # Prevent indefinite recursion.
    old_cycle_len = len(cycle)
    old_guess_based = guess_based
    guess_based = False
    old_recursion_hits = recursion_hits   
    # We set the resolution to the value that is returned when
    # this order is visited deeper in recursion.
    order.resolution = False
    first_result = adjudicate(order)

    if not guess_based:
        # No cyclic dependencies were detected.
        # We can just take this resolution.
        guess_based = old_guess_based
        order.resolution = first_result
        order.resolved = True
        return first_result

    # If order in cycle, this order was hit in recursion, but
    # we might need to retreat further in recursion.
    if order in cycle:
        recursion_hits -= 1

    if recursion_hits == old_recursion_hits:
        # Deeper in the recursion we hit a cycle on this order.
        # And this order is the ancestor of the whole cycle.
        # Try to adjudicate the cycle with different guess.    
        del cycle[old_cycle_len:]
        order.resolution = True # Was False on first run.
        second_result = adjudicate(order)

        if first_result == second_result:
            # Although we hit a cycle, there is only one result.
            del cycle[old_cycle_len:]
            guess_based = old_guess_based
            order.resolution = first_result
            order.resolved = True
            return first_result

        # Different results on different guesses.
        backup_rule(cycle[old_cycle_len:])
        del cycle[old_cycle_len:]
        guess_based = old_guess_based
        order.visited = False
        # Backup rule might not have resolved this order.
        return resolve(order)

    # We are returning from recursion where we hit a cycle,
    # but this order is not the ancestor of the whole cycle.
    if not order in cycle:
        cycle.add(order)
    # We remember the result if resolve is called again on same
    # order, while cycle is not yet resolved.
    order.resolution = first_result
    order.visited = False
    return first_result
